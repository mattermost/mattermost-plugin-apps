const util = require('util');

exports.handler = async (event) => {
    console.log(util.inspect(event, {showHidden: false, depth: null}));

    return {
        'Type': 'ok',
        'Markdown': 'hello world',
    };;
};