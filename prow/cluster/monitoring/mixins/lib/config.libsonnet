{
  _config+:: {
    // Grafana dashboard IDs are necessary for stable links for dashboards
    grafanaDashboardIDs: {
      'ghproxy.json': 'd72fe8d0400b2912e319b1e95d0ab1b3',
      'slo.json': 'ea313af4b7904c7c983d20d9572235a5',
    },
    // Component name constants
    components: {
      // Values should be lowercase for use with prometheus 'job' label.
      crier: 'crier',
      deck: 'deck',
      ghproxy: 'ghproxy',
      hook: 'hook',
      horologium: 'horologium',
      monitoring: 'monitoring', // Aggregate of prometheus, alertmanager, and grafana.
      plank: 'plank', // Mutually exclusive with prowControllerManager
      prowControllerManager: 'prow-controller-manager',
      sinker: 'sinker',
      tide: 'tide',
    },
    local comps = self.components,

    // SLO compliance tracking config
    slo: {
      components: [
        comps.deck,
        comps.hook,
        comps.plank,
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
    prowImageStaleByDays: 14,
  },
}
