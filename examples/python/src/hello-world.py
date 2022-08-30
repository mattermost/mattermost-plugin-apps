import logging
import os
from posixpath import join

import requests
from flask import Flask, request

logging.basicConfig(level=logging.DEBUG)

app = Flask(__name__, static_url_path='/static', static_folder='./static')

default_port = 8080
default_host = 'localhost'
default_root_url = 'http://localhost:8080'
SHARED_FORM = {
    'title': 'I am a form!',
    'icon': 'icon.png',
    'fields': [
        {
            'type': 'text',
            'name': 'message',
            'label': 'message',
            'position': 1,
        }
    ],
    'submit': {
        'path': '/submit',
    },
}


@app.route('/manifest.json')
def manifest() -> dict:
    return {
        'app_id': 'hello-world',
        'display_name': 'Hello world app',
        'homepage_url': 'https://github.com/mattermost/mattermost-plugin-apps/tree/master/examples/python/hello-world',
        'app_type': 'http',
        'icon': 'icon.png',
        'requested_permissions': ['act_as_bot'],
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
        'root_url': os.environ.get('ROOT_URL', default_root_url),
    }


@app.route('/submit', methods=['POST'])
def on_form_submit():
    print(request.json)
    return {'type': 'ok', 'text': f'Hello form got submitted. Form data: {request.json["values"]}'}


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
                        'icon': 'icon.png',
                        'submit': {
                            'path': '/first_command',
                            # expand block is optional. This is more of metadata like which channel, team this command
                            # was called from
                            'expand': {
                                'app': 'all',
                                # if you want to expand team & channel, ensure that bot is added to the team & channel
                                # else command will fail to expand the context
                                # 'team': 'all',
                                # 'channel': 'all',
                            },
                        },
                    },
                    {   # command with embedded form
                        'description': 'test command',
                        'hint': '[This is testing command]',
                        # this will be the command displayed to user as /second-command
                        'label': 'second-command',
                        'icon': 'icon.png',
                        'bindings': [
                            {
                                # sub-command `send` to send an embedded form here as input to the command.
                                # E.g. /second-command send "hello-form"
                                'location': 'send',
                                'label': 'send',
                                'form': SHARED_FORM
                            },
                        ],
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
                        'form': SHARED_FORM,
                    },
                ],
            },
        ],
    }


@app.route('/ping', methods=['POST'])
def on_ping() -> dict:
    logging.debug('ping...')
    return {'type': 'ok'}


@app.route('/install', methods=['GET', 'POST'])
def on_install() -> dict:
    print(f'on_install called with payload , {request.args}, {request.data}', flush=True)
    _subscribe_team_join(request.json['context'])
    return {'type': 'ok', 'data': []}


@app.route('/first_command', methods=['POST'])
def on_first_command():
    print(f'/first_command called ')
    response_message = 'Hello! response from /first_command'
    return {'type': 'ok', 'text': response_message}


@app.route('/bot_joined_team', methods=['GET', 'POST'])
def on_bot_joined_team() -> dict:
    context = request.json['context']
    logging.info(
        f'bot_joined_team event received for site:{context["mattermost_site_url"]}, '
        f'team:{context["team"]["id"]} name:{context["team"]["name"]} '
        f'{request.args} {request.data}'
    )
    # Here one can subscribe to channel_joined/left events as these required team_id now to be subscribed,
    # hence use the team_id received in the event and make a call for subscribing to channel_joined/left events.
    # Also supply {'team_id': team_id} in the request body of the subscription
    # {
    #    'subject': 'bot_joined_team',
    #    'call': {
    #        'path': '/bot_joined_team',
    #         'expand': {
    #             'app': 'all',
    #             'team': 'all'
    #         }
    #    },
    #    'team_id': 'team_id'   # get this team_id when bot_joined_team event occurs
    # }
    return {'type': 'ok', 'data': []}


# Subscribing to events. For example, Subscribe to 'bot_joined_team' event
def _subscribe_team_join(context: dict) -> None:
    site_url = context['mattermost_site_url']
    bot_access_token = context['bot_access_token']
    url = join(site_url, 'plugins/com.mattermost.apps/api/v1/subscribe')
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
    app.run(
        debug=True,
        host=os.environ.get('HOST', default_host),
        port=int(os.environ.get('PORT', default_port)),
        use_reloader=False,
    )
