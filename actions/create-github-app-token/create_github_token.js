module.exports = ({ audience }) => {
    async function retry(fn, retries = 3, delay = 5000) {
        for (let i = 0; i < retries; i++) {
            try {
                return await fn();
            } catch (err) {
                core.warning(`Attempt ${i + 1} failed: ${err.message}`);
                if (i < retries - 1) {
                    await new Promise(r => setTimeout(r, delay * i));
                } else {
                    throw err;
                }
            }
        }
    }
    const jwt = retry(() => core.getIDToken(audience));
    core.setSecret(jwt);
    core.setOutput("github-jwt", jwt);
}
