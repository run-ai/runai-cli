{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "runai.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "runai-common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "runai-common.charts.label-addition"}}
app: {{ template "runai.name" . }}
chart: {{ template "runai-common.chart" . }}
release: {{ .Release.Name }}
heritage: {{ .Release.Service }}
createdBy: "RunaiJob"
{{- end }}

{{- define "host.path.volume.name" -}}
{{- $volumeIndex := (get . "volumeIndex") -}}
{{ printf "host-path-volume-%d" $volumeIndex }}
{{- end -}}

{{- define "pvc.volume.name" -}}
{{- $volumeIndex := (get . "volumeIndex") -}}
{{ printf "pvc-volume-%d" $volumeIndex }}
{{- end -}}

{{- define "pvc.claim.name" -}}
{{- $pvcIndex := (get . "pvcIndex") -}}
{{- $releaseName := (get . "releaseName") -}}
{{ printf "pvc-%s-%d" $releaseName $pvcIndex }}
{{- end -}}

