import https from 'node:https';
import { URL } from 'node:url';
import { ProxyAgent } from 'proxy-agent';
const agent = new ProxyAgent();
export async function postData(urlString, data, headers) {
    return new Promise((resolve, reject) => {
        const url = new URL(urlString);
        const postData = JSON.stringify(data);
        const options = {
            hostname: url.hostname,
            port: url.port || 443,
            path: url.pathname,
            method: 'POST',
            headers: {
                ...headers,
                'Content-Type': 'application/json',
            },
            agent,
        };
        const req = https.request(options, (res) => {
            const chunks = [];
            res.on('data', (chunk) => chunks.push(chunk));
            res.on('end', () => {
                const results = Buffer.concat(chunks);
                resolve({
                    data: results.toString(),
                    status: res.statusCode ?? 200,
                });
            });
            res.on('error', reject);
        });
        req.on('error', reject);
        req.write(postData);
        req.end();
    });
}
