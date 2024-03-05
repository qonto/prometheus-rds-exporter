local g = import 'g.libsonnet';
local link = g.dashboard.link;
local options = link.dashboards.options;

{
  uuids: {
    rdsInstance: 'a7049b32-6be3-42e5-aa9a-2879a14f46dd',
    rdsInstances: 'efa71e45-3356-4206-b61f-1e2a3e4e2185',
    prometheusRDSExporter: 'f65d785e-d8c2-49b5-8314-388f30483f57',
  },
  colors: {
    transparent: 'transparent',
    default: 'white',
    ok: 'green',
    notice: 'yellow',
    warning: 'orange',
    danger: 'red',
    limit: 'red',
  },
  links: [
    link.dashboards.new('Server', ['dmf', 'server'])
    + options.withAsDropdown(true)
    + options.withIncludeVars(true)
    + options.withKeepTime(true),
    link.dashboards.new('Database', ['dmf', 'database'])
    + options.withAsDropdown(true)
    + options.withIncludeVars(true)
    + options.withKeepTime(true),
    link.dashboards.new('Table', ['dmf', 'table'])
    + options.withAsDropdown(true)
    + options.withIncludeVars(true)
    + options.withKeepTime(true),
    link.dashboards.new('More', ['dmf', 'misc'])
    + options.withAsDropdown(true)
    + options.withIncludeVars(true)
    + options.withKeepTime(true),
  ],
}
