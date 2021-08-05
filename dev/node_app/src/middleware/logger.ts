export default (req, res, next) => {
    const headers = req.headers ? stringify(req.headers) : 'no headers';

    const body = req.body ? stringify(req.body) : 'no body';

    const output = `\n${req.method} ${req.url}\nHeaders: ${headers}\nRequest: ${body}`;
    console.log(output);

    // Hook into send method to log response
    const send = res.send.bind(res);
    res.send = (...args) => {
        let output = args[0];
        try {
            const jsonBody = JSON.parse(output);
            output = stringify(jsonBody);
        } catch (e) {}

        console.log(`Response: ${res.statusCode} ${output}`);

        // Call actual send method
        send(...args);
    }

    next();
}

const stringify = (data) => {
    return JSON.stringify(data, null, 2);
}
