{{/* vim: set filetype=mustache: */}}

{{/* Generate basic labels */}}
{{- define "chart.labels" }}
labels:
  {{include "runai-common.charts.label-addition" . | indent 2}}
  app: {{ template "runai.name" . }}
  chart: {{ template "runai-common.chart" . }}
  release: {{ .Release.Name }}
  heritage: {{ .Release.Service }}
  createdBy: "RunaiJob"
{{- end }}