const util = require('util');
// Create a DocumentClient that represents the query to add an item
const dynamodb = require('aws-sdk/clients/dynamodb');
const docClient = new dynamodb.DocumentClient();
const apiConnectStartURL = 'https://m20ldarqw9.execute-api.us-east-2.amazonaws.com/dev';
const apiConnectFinishURL = '';
console.log('Loading install function');

exports.installHandler = async (event) => {
    console.log(util.inspect(event, {showHidden: false, depth: null}));

}

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