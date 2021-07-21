// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upkubeless

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-getter"
	kubelessAPI "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubelessutil "github.com/kubeless/kubeless/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// ProvisionApp creates all necessary functions in Kubeless, and outputs
// the manifest to use.
//
// Its input is a zip file containing:
//   |-- manifest.json
//   |-- function files referenced in manifest.json...
func ProvisionApp(bundlePath string, log Logger, shouldUpdate bool) (*apps.Manifest, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp directory to unpack the bundle")
	}
	defer os.RemoveAll(dir)

	pwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain current working directory")
	}

	getBundle := getter.Client{
		Mode: getter.ClientModeDir,
		Src:  bundlePath,
		Dst:  dir,
		Pwd:  pwd,
		Ctx:  context.Background(),
	}
	err = getBundle.Get()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bundle "+bundlePath)
	}

	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load manifest.json")
	}
	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return nil, errors.Wrap(err, "invalid manifest.json")
	}
	if log != nil {
		log.Info("Loaded App bundle", "bundle", bundlePath, "app_id", m.AppID)
	}

	k8sClient := kubelessutil.GetClientOutOfCluster()
	ns := namespace(m.AppID)
	existing, _ := k8sClient.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
	if existing == nil {
		_, err := k8sClient.CoreV1().Namespaces().Create(
			&v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: ns,
				},
			})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create kubernetes namespace %s", ns)
		}
	}

	kubelessClient, err := kubelessutil.GetKubelessClientOutCluster()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Kubeless client")
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
				Namespace: ns,
				Name:      fName,
			},
			Spec: kubelessAPI.FunctionSpec{
				Handler: kf.File + "." + kf.Handler,
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

		fmt.Printf("<>/<> 1 %s\n", utils.Pretty(f))

		if shouldUpdate {
			_, err = kubelessutil.GetFunctionCustomResource(kubelessClient, f.Name, f.Namespace)
			fmt.Printf("<>/<> 2 %v\n", err)
			switch {
			case err == nil:
				err = kubelessutil.PatchFunctionCustomResource(kubelessClient, f)
				fmt.Printf("<>/<> 3 %v\n", err)
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
		fmt.Printf("<>/<> 4 %v\n", err)
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
