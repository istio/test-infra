{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'prow-stale',
        rules: [
          {
            alert: 'Prow images are stale',
            expr: |||
              time()-max(prow_version) > %d * 24 * 3600
            ||| % $._config.prowImageStaleByDays,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              message: 'The prow images are older than %d days.' % $._config.prowImageStaleByDays,
            },
          }
        ],
      },
    ],
  },
}
