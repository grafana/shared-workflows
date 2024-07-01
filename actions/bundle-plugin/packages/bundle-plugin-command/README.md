# Grafana / Zip Plugin

Zip Grafana plugins for distribution.

**`@grafana/zip-bundle-plugin`** works on macOS, Windows and Linux.<br /> If
something doesn’t work, please
[file an issue](https://github.com/grafana/shared-workflows/issues/new).<br />
If you have questions or need help, please ask in
[GitHub Discussions](https://github.com/grafana/shared-workflows/discussions).

## Packaging a plugin

Packaging a plugin in a zip file is the recommended way of distributing Grafana
plugins.

Ensure the following environment variables are set before running the script:

- `GRAFANA_ACCESS_POLICY_TOKEN`: Used to sign plugins.
  [Generate a token](https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin#generate-an-access-policy-token).

- `GRAFANA_API_KEY`: Deprecated. Consider using `GRAFANA_ACCESS_POLICY_TOKEN`
  instead.

If neither is set, use the `--noSign` flag to skip plugin signing. This will
result in an unsigned bundle that cannot be validated or uploaded for official
use, but may be good for testing.

If your plugin puts build output in /dist, you can just do:

```bash
npx @grafana/bundle-plugin@latest
```

If the plugin distribution directory differs from the default `dist`, specify
the path to use with the `--distDir` flag.

```bash
npx @grafana/bundle-plugin@latest --distDir path/to/directory
```

Alternatives:

#### [`github actions`](https://docs.github.com/en/actions)

```yaml
npx @grafana/bundle-plugin@latest
```

#### [`yarn`](https://yarnpkg.com/cli/dlx) (> 2.x)

```bash
yarn dlx @grafana/bundle-plugin@latest
```

## Contributing

We are always grateful for contribution! See the
[CONTRIBUTING.md](../../CONTRIBUTING.md) for more information.
