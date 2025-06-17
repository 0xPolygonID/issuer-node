{{/*
Common labels
*/}}
{{- define "privadoid-issuer.postgresIssuerNode.common.labels" -}}
helm.sh/chart: {{ include "privadoid-issuer.chart" . }}
{{ include "privadoid-issuer.postgresIssuerNode.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "privadoid-issuer.postgresIssuerNode.staticLabel" -}}
app: {{ .Values.postgresIssuerNode.labels.app }}
{{- end }}

{{- define "privadoid-issuer.postgresIssuerNode.selectorLabels" -}}
app.kubernetes.io/name: {{ .Release.Name }}
{{- end }}

{{/*
Define custom deployment selectorLabels for postgres
*/}}
{{- define "privadoid-issuer.postgresIssuerNode.deploymentLabels" -}}
app: {{ .Values.postgresIssuerNode.deployment.labels.app }}
{{- end }}


{{/*
Define custom service selectorLabels for postgres
*/}}
{{- define "privadoid-issuer.postgresIssuerNode.Labels" -}}
app: {{ .Values.postgresIssuerNode.service.selector }}
{{- end }}
