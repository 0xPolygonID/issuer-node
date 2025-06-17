{{/*
Common labels
*/}}
{{- define "privadoid-issuer.redisIssuerNode.common.labels" -}}
helm.sh/chart: {{ include "privadoid-issuer.chart" . }}
{{ include "privadoid-issuer.redisIssuerNode.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Define custom service selectorLabels for redis
*/}}
{{- define "privadoid-issuer.redisIssuerNode.Labels" -}}
app: {{ .Values.redisIssuerNode.service.selector }}
{{- end }}

{{- define "privadoid-issuer.redisIssuerNode.selectorLabels" -}}
app.kubernetes.io/name: {{ .Release.Name }}
{{- end }}

{{- define "privadoid-issuer.redisIssuerNode.staticLabel" -}}
app: {{ .Values.redisIssuerNode.labels.app }}
{{- end }}