{{/*
Expand the name of the chart.
*/}}
{{- define "server-room.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "server-room.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "server-room.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "server-room.labels" -}}
helm.sh/chart: {{ include "server-room.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
NATS labels
*/}}
{{- define "server-room.natsLabels" -}}
{{ include "server-room.labels" . }}
app.kubernetes.io/name: server-room-nats
app.kubernetes.io/component: nats
{{- end }}

{{/*
NATS selector labels
*/}}
{{- define "server-room.natsSelectorLabels" -}}
app.kubernetes.io/name: server-room-nats
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Dashboard labels
*/}}
{{- define "server-room.dashboardLabels" -}}
{{ include "server-room.labels" . }}
app.kubernetes.io/name: server-room-dashboard
app.kubernetes.io/component: dashboard
{{- end }}

{{/*
Dashboard selector labels
*/}}
{{- define "server-room.dashboardSelectorLabels" -}}
app.kubernetes.io/name: server-room-dashboard
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
