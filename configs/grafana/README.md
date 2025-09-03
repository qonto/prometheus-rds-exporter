# Grafana Dashboards for Prometheus RDS Exporter

This folder contains the configuration and tools needed to build and publish Grafana dashboards for the Prometheus RDS Exporter. The dashboards are defined using [Jsonnet](https://jsonnet.org/) and [Grafonnet](https://github.com/grafana/grafonnet), which allows for programmatic dashboard creation.

## Directory Structure

- `dashboards/`: Contains Jsonnet dashboard definitions
- `panels/`: Reusable panel components
- `queries/`: Prometheus query templates
- `public/`: Generated JSON dashboard files (output)
- `vendor/`: Dependencies installed by jsonnet-bundler
- `build.sh`: Script to build dashboards
- `publish.sh`: Script to publish dashboards to a Grafana instance

## Requirements

To build and publish these dashboards, you need:

1. **Jsonnet**: A data templating language

   ```
   # macOS
   brew install jsonnet
   
   # Linux
   apt-get install jsonnet
   ```

2. **jsonnet-bundler (jb)**: A package manager for Jsonnet

   ```
   go install github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@latest
   ```

3. **jq**: A lightweight JSON processor

   ```
   # macOS
   brew install jq
   
   # Linux
   apt-get install jq
   ```

4. **curl**: For API requests to Grafana

   ```
   # macOS
   brew install curl
   
   # Linux
   apt-get install curl
   ```

5. **Grafana instance**: Running Grafana server for dashboard publishing

   We recommand to use our local development environment as configuration/credentials are pre-defined

6. Optional. Install VScode [Jsonnet Language Server extension](https://marketplace.visualstudio.com/items?itemName=Grafana.vscode-jsonnet) to faciliate dashboard development

## Building Dashboards

To build all dashboards:

```bash
./build.sh
```

To build a specific dashboard:

```bash
./build.sh dashboard_name
```

This will generate JSON files in the `public/` directory.

## Publishing Dashboards

Note: publish is only used for local development, works out-of-the box with local development environment

To publish all dashboards to Grafana:

```bash
./publish.sh
```

To publish a specific dashboard:

```bash
./publish.sh dashboard_name
```

### Environment Variables for Publishing

You can customize the Grafana connection using these environment variables:

- `GRAFANA_URL`: Grafana server URL (default: <http://localhost:3000>)
- `GRAFANA_USERNAME`: Grafana username (default: admin)
- `GRAFANA_PASSWORD`: Grafana password (default: hackme)

Example:

```bash
GRAFANA_URL=http://grafana.example.com GRAFANA_USERNAME=user GRAFANA_PASSWORD=pass ./publish.sh
```

## Creating New Dashboards

1. Create a new Jsonnet file in the `dashboards/` directory
2. Create a new dashboard UUID in `common.libsonnet`
3. Use the existing dashboards as templates
4. Update `dashboard.withUid()` in template definiting using the new UUID
5. Leverage the common components in `panels/` and `queries/`
6. Start local development environment
7. Build and test your dashboard using `./build.sh && ./publish.sh`
8. Submit a pull request

## Dependencies

This project uses [Grafonnet](https://github.com/grafana/grafonnet) for dashboard generation, which is automatically installed by jsonnet-bundler when running the build script.
