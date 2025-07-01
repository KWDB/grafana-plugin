# KWDB Data Source

KWDB Data Source is a plugin for Grafana that allows you to query and visualize data from KWDB.

<!-- To help maximize the impact of your README and improve usability for users, we propose the following loose structure:

**BEFORE YOU BEGIN**
- Ensure all links are absolute URLs so that they will work when the README is displayed within Grafana and Grafana.com
- Be inspired ✨
  - [grafana-polystat-panel](https://github.com/grafana/grafana-polystat-panel)
  - [volkovlabs-variable-panel](https://github.com/volkovlabs/volkovlabs-variable-panel)

**ADD SOME BADGES**

Badges convey useful information at a glance for users whether in the Catalog or viewing the source code. You can use the generator on [Shields.io](https://shields.io/badges/dynamic-json-badge) together with the Grafana.com API
to create dynamic badges that update automatically when you publish a new version to the marketplace.

- For the URL parameter use `https://grafana.com/api/plugins/your-plugin-id`.
- Example queries:
  - Downloads: `$.downloads`
  - Catalog Version: `$.version`
  - Grafana Dependency: `$.grafanaDependency`
  - Signature Type: `$.versionSignatureType`
- Optionally, for the logo parameter use `grafana`.

Full example: ![Dynamic JSON Badge](https://img.shields.io/badge/dynamic/json?logo=grafana&query=$.version&url=https://grafana.com/api/plugins/grafana-polystat-panel&label=Marketplace&prefix=v&color=F47A20)

Consider other [badges](https://shields.io/badges) as you feel appropriate for your project. -->

## Adding the data source

To add the KWDB data source to Grafana, follow these steps:

1. In Grafana, go to **Connection** > **Add new connection**.
2. Select **KWDB** from the list of data sources.
3. Click **Add new data source**.
4. Enter the name of the data source.
5. Enter the Connection and Authentication info of the KWDB instance.
6. Click **Save & Test**.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.
