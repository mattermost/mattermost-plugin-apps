import logging
import os

import requests
from flask import Flask, request

logging.basicConfig(level=logging.DEBUG)

APP_PORT = 8080
APP_HOST = 'localhost'
ROOT_URL = 'http://localhost:8080'

app = Flask(__name__)


@app.route('/manifest.json')
def manifest() -> dict:
    return {
        'app_id': 'hello-world',
        'display_name': 'Hello world app',
        'homepage_url': 'https://github.com/mattermost/mattermost-plugin-apps/tree/master/examples/python/hello-world',
        'app_type': 'http',
        'requested_permissions': [
            'act_as_bot'
        ],
        'on_install': {
            'path': '/install',
            'expand': {
                'app': 'all',
            },
        },
        'bindings': {
            'path': '/bindings',
        },
        'requested_locations': [
            '/channel_header',
            '/command'
        ],
        'root_url': ROOT_URL,
    }


@app.route('/bindings', methods=['GET', 'POST'])
def on_bindings() -> dict:
    print(f'bindings called with {request.data}')
    return {
        'type': 'ok',
        'data': [
            {
                # binding for a command
                'location': '/command',
                'bindings': [
                    {
                        'description': 'test command',
                        'hint': '[This is testing command]',
                        # this will be the command displayed to user as /first-command
                        'label': 'first-command',
                        'submit': {
                            'path': '/first_command',
                            'expand': {
                                'app': 'all',
                                'team': 'all',
                                'channel': 'all',
                            },
                        },
                    }
                ],
            },
            {
                'location': '/channel_header',
                'bindings': [
                    {
                        'location': 'send-button',
                        'icon': 'icon.png',
                        'label': 'send hello message',
                        'call': {
                            'path': '/send-modal',
                        },
                    },
                ],
            }
        ],
    }


@app.route('/install', methods=['GET', 'POST'])
def on_install() -> dict:
    print(f'on_install called with payload , {request.args}, {request.data}', flush=True)
    _subscribe_team_join(request.json['context'])
    return {'type': 'ok', 'data': []}


@app.route('/first_command', methods=['POST'])
def on_first_command():
    print(f'/first_command called ')
    return {'type': 'ok'}


@app.route('/bot_joined_team', methods=['GET', 'POST'])
def on_bot_joined_team() -> dict:
    context = request.json['context']
    logging.info(
        f'bot_joined_team event received for site:{context["mattermost_site_url"]}, '
        f'team:{context["team"]["id"]} name:{context["team"]["name"]} '
        f'{request.args} {request.data}'
    )
    return {'type': 'ok', 'data': []}


"""
Subscribing to events. For example, Subscribe to 'bot_joined_team' event
"""

def _subscribe_team_join(context: dict) -> None:
    site_url = context['mattermost_site_url']
    bot_access_token = context['bot_access_token']
    url = os.path.join(site_url, 'plugins/com.mattermost.apps/api/v1/subscribe')
    logging.info(f'Subscribing to team_join for {site_url}')
    headers = {'Authorization': f'BEARER {bot_access_token}'}
    body = {
        'subject': 'bot_joined_team',
        'call': {
            'path': '/bot_joined_team',
            'expand': {
                'app': 'all',
                'team': 'all'
            }
        },
    }
    res = requests.post(url, headers=headers, json=body)
    if res.status_code != 200:
        logging.error(f'Could not subscribe to team_join event for {site_url}')
    else:
        logging.debug(f'subscribed to team_join event for {site_url}')


if __name__ == '__main__':
    app.run(debug=True, host=APP_HOST, port=int(APP_PORT), use_reloader=False)
