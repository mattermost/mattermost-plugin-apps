const util = require('util');
// Create a DocumentClient that represents the query to add an item
const dynamodb = require('aws-sdk/clients/dynamodb');
const docClient = new dynamodb.DocumentClient();
const apiConnectStartURL = 'https://m20ldarqw9.execute-api.us-east-2.amazonaws.com/dev';

exports.installHandler = async (event) => {
    console.log(util.inspect(event, {showHidden: false, depth: null}))

    let appData = {
        'bot_access_token' : event.values.data.bot_access_token,
        'oauth2_client_secret' : event.values.data.oauth2_client_secret,
        'bot_user_id': event.context.app.bot_user_id,
        'oauth2_client_id': event.context.app.oauth2_client_id
    };
    let params = {
        TableName: event.context.app_id,
        Item: { appData },
    };

    await docClient.put(params).promise();

    //store context for user
    params = {
        TableName: event.context.acting_user_id,
        Item: event.context,
    };
    await docClient.put(params).promise();

    const respText = `**Hallo სამყარო** needs to continue its installation using your system administrator's credentials. Please [connect](${apiConnectStartURL}) the application to your Mattermost account.`;
    const response = {
        'Type': 'ok',
        'Markdown': respText,
    };
    return response;
};

var startOAuth2Connect = function (context) {
    // store context

    const state = 'some_random_string';
    // store state

    const url = new URL(apiConnectStartURL)
    url.append('response_type', 'code');
    url.append('client_id', context.app.oauth2_client_id);
    url.append('redirect_uri', apiConnectFinishURL);
    url.append('state', state);
    url.append('access_type', 'offline');

    return url.toString();
}