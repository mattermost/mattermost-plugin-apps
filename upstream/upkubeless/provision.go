// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upkubeless

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-getter"
	kubelessAPI "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kubelessutil "github.com/kubeless/kubeless/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// ProvisionApp creates all necessary functions in Kubeless, and outputs
// the manifest to use.
//
// Its input is a zip file containing:
//   |-- manifest.json
//   |-- function files referenced in manifest.json...
func ProvisionApp(kubelessClient versioned.Interface, bundlePath string, log Logger, shouldUpdate bool) (*apps.Manifest, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp directory to unpack the bundle")
	}
	defer os.RemoveAll(dir)

	err = getter.Get(dir, bundlePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bundle "+bundlePath)
	}

	// Load manifest.json
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load manifest.json")
	}
	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return nil, errors.Wrap(err, "invalid manifest.json")
	}

	// Provision functions.
	for _, kf := range m.KubelessFunctions {
		fName := FunctionName(m.AppID, m.Version, kf.CallPath)

		f := &kubelessAPI.Function{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Function",
				APIVersion: "kubeless.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"created-by": "Mattermost Kubeless Apps",
					"function":   fName,
				},
				Namespace: namespace(m.AppID),
				Name:      fName,
			},
			Spec: kubelessAPI.FunctionSpec{
				Handler: kf.File + "." + kf.Name,
				Runtime: kf.Runtime,
				Timeout: kf.Timeout,
				ServiceSpec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:     "http-function-port",
							Protocol: v1.ProtocolTCP,
						},
					},
					Type: v1.ServiceTypeClusterIP,
				},
			},
		}

		f.Spec.FunctionContentType, f.Spec.Function, err = loadFile(dir, kf.File, kf.Checksum)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load "+kf.File)
		}
		f.Spec.Checksum = kf.Checksum

		if kf.DepsFile != "" {
			_, f.Spec.Deps, err = loadFile(dir, kf.DepsFile, "")
			if err != nil {
				return nil, errors.Wrap(err, "failed to load "+kf.DepsFile)
			}
		}

		if shouldUpdate {
			_, err = kubelessutil.GetFunctionCustomResource(kubelessClient, f.Name, f.Namespace)
			switch {
			case err == nil:
				err = kubelessutil.PatchFunctionCustomResource(kubelessClient, f)
				if err != nil {
					return nil, errors.Wrap(err, "failed to patch function "+f.Name)
				}
				return m, nil

			case k8sErrors.IsNotFound(err):
				// Fall to create the function.

			default:
				return nil, errors.Wrap(err, "failed to get function "+f.Name)
			}
		}

		err = kubelessutil.CreateFunctionCustomResource(kubelessClient, f)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create function "+f.Name)
		}
	}

	return m, nil
}

func loadFile(dir, name, checksum string) (string, string, error) {
	fileName := filepath.Join(dir, name)
	contentType, err := kubelessutil.GetContentType(fileName)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to detect content type "+name)
	}
	data, actualChecksum, err := kubelessutil.ParseContent(fileName, contentType)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to load "+name)
	}
	if checksum != "" && checksum != actualChecksum {
		return "", "", errors.Wrapf(err, "checksum mismatch for %s, expected %s, got %s", name, checksum, actualChecksum)
	}
	return contentType, data, nil
}
