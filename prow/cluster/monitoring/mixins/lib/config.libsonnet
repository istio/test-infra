local util = import 'config_util.libsonnet';

//
// Edit configuration in this object.
//
local config = {
  local comps = util.consts.components,

  // Instance specifics
  instance: {
    name: "Istio Prow",
    botName: "istio-testing",
    url: "https://prow.istio.io",
    monitoringURL: "https://monitoring.prow.istio.io",
  },

  // SLO compliance tracking config
  slo: {
    components: [
      comps.deck,
      comps.hook,
      comps.prowControllerManager,
      comps.sinker,
      comps.tide,
      comps.monitoring,
    ],
  },

  // Tide pools that are important enough to have their own graphs on the dashboard.
  tideDashboardExplicitPools: [],

  // Additional scraping endpoints
  probeTargets: [
    # ATTENTION: Keep this in sync with the list in ../../additional-scrape-configs_secret.yaml
    {url: 'https://prow.istio.io', labels: {slo: comps.deck}},
    {url: 'https://monitoring.prow.istio.io', labels: {}},
  ],

  // Boskos endpoints to be monitored
  boskosResourcetypes: [],

  // How long we go during work hours without seeing a webhook before alerting.
  webhookMissingAlertInterval: '30m',

  // How many days prow hasn't been bumped.
  prowImageStaleByDays: {daysStale: 7, eventDuration: '24h'},
};

// Generate the real config by adding in constant fields and defaulting where needed.
{
  _config+:: util.defaultConfig(config),
  _util+:: util,
}
