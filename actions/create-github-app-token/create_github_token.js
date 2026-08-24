module.exports = ({ core, audience }) => {
  async function retry(fn, retries = 3, delay = 5000) {
    console.log(`Audience: ${audience}`);
    for (let i = 0; i < retries; i++) {
      console.log(`Attempt ${i + 1}`);
      try {
        return await fn();
      } catch (err) {
        core.warning(`Attempt ${i + 1} failed: ${err.message}`);
        if (i < retries - 1) {
          await new Promise((r) => setTimeout(r, delay * i));
        } else {
          throw err;
        }
      }
    }
  }

  return retry(() => core.getIDToken(audience))
    .then((jwt) => {
      core.setSecret(jwt);
      core.setOutput("github-jwt", jwt);
    })
    .catch((err) => {
      core.setFailed(`Failed to get ID token: ${err.message}`);
      throw err; // ensure caller knows it failed
    });
};
