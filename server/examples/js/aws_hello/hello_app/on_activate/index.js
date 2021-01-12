const util = require('util');

exports.handler = async (event) => {
    requestText = util.inspect(event, {showHidden: false, depth: null});
    console.log(requestText);

    return {
        'Type': 'ok',
        'Markdown': '```\n' + requestText + '```\n',
    };;
};