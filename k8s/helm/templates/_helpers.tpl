
{{/*
Expand the name of the chart.
*/}}
{{- define "polygon-id-issuer.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "polygon-id-issuer.fullname" -}}
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
{{- define "polygon-id-issuer.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "polygon-id-issuer.labels" -}}
helm.sh/chart: {{ include "polygon-id-issuer.chart" . }}
{{ include "polygon-id-issuer.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "polygon-id-issuer.selectorLabels" -}}
app.kubernetes.io/name: {{ .Release.Name }}
{{- end }}



{{/*
Define a static label 
*/}}
{{- define "polygon-id-issuer.staticLabel" -}}
app: {{ .Values.apiIssuerNode.service.labels.app }}
{{- end }}


{{/*
Define an env var
*/}}
{{- define "helpers.issuer-db-url" -}}
ISSUER_DATABASE_URL
{{- end }}

{{/*
Define an env var
*/}}
{{- define "helpers.issuer-key-store-addr" -}}
ISSUER_KEY_STORE_ADDRESS
{{- end }}

{{/*
Define custom service selectorLabels for apiIssuerNode
*/}}
{{- define "polygon-id-issuer.apiIssuerNode.Labels" -}}
app: {{ .Values.apiIssuerNode.service.selector }}
{{- end }}

{{/*
Define custom deployment labels fors apiIssuerNode
*/}}
{{- define "polygon-id-issuer.apiIssuerNode.deploymentLabels" -}}
app: {{ .Values.apiIssuerNode.deployment.labels.app }}
{{- end }}

{{/*
Define custom service selectorLabels for apiUiIssuerNode
*/}}
{{- define "polygon-id-issuer.apiUiIssuerNode.Labels" -}}
app: {{ .Values.apiUiIssuerNode.service.selector }}
{{- end }}

{{/*
Define custom deployment selectorLabels for apiUiIssuerNode
*/}}
{{- define "polygon-id-issuer.apiUiIssuerNode.deploymentLabels" -}}
app: {{ .Values.apiUiIssuerNode.deployment.labels.app }}
{{- end }}


{{/*
Define custom deployment selectorLabels for notifications
*/}}
{{- define "polygon-id-issuer.notificationsIssuerNode.Labels" -}}
app: {{ .Values.notificationsIssuerNode.deployment.name }}
{{- end }}

{{/*
Define custom deployment label for notifications
*/}}
{{- define "polygon-id-issuer.notificationsIssuerNode.deploymentLabels" -}}
app: {{ .Values.notificationsIssuerNode.deployment.labels.app }}
{{- end }}

{{/*
Define custom deployment selectorLabels for pending-publisher
*/}}
{{- define "polygon-id-issuer.pendingPublisherIssuerNode.Labels" -}}
app: {{ .Values.pendingPublisherIssuerNode.deployment.name }}
{{- end }}

{{/*
Define custom deployment label for pending-publisher
*/}}
{{- define "polygon-id-issuer.pendingPublisherIssuerNode.deploymentLabels" -}}
app: {{ .Values.pendingPublisherIssuerNode.deployment.labels.app }}
{{- end }}


{{/*
Define custom service selectorLabels for UiIssuerNode
*/}}
{{- define "polygon-id-issuer.uiIssuerNode.Labels" -}}
app: {{ .Values.uiIssuerNode.service.selector }}
{{- end }}


{{/*
Define custom deployment selectorLabels for UiIssuerNode
*/}}
{{- define "polygon-id-issuer.uiIssuerNode.deploymentLabels" -}}
app: {{ .Values.uiIssuerNode.deployment.labels.app }}
{{- end }}

{{/*
Define custom service selectorLabels for postgres
*/}}
{{- define "polygon-id-issuer.postgresIssuerNode.Labels" -}}
app: {{ .Values.postgresIssuerNode.service.selector }}
{{- end }}


{{/*
Define custom deployment selectorLabels for postgres
*/}}
{{- define "polygon-id-issuer.postgresIssuerNode.deploymentLabels" -}}
app: {{ .Values.postgresIssuerNode.deployment.labels.app }}
{{- end }}


{{/*
Define custom service selectorLabels for redis
*/}}
{{- define "polygon-id-issuer.redisIssuerNode.Labels" -}}
app: {{ .Values.redisIssuerNode.service.selector }}
{{- end }}


{{/*
Define custom deployment selectorLabels for vault
*/}}
{{- define "polygon-id-issuer.vaultIssuerNode.deploymentLabels" -}}
app: {{ .Values.vaultIssuerNode.deployment.labels.app }}
{{- end }}

{{/*
Define custom service selectorLabels for vault
*/}}
{{- define "polygon-id-issuer.vaultIssuerNode.Labels" -}}
app: {{ .Values.vaultIssuerNode.service.selector }}
{{- end }}
