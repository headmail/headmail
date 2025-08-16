{{/* Common helpers for headmail chart */}}
{{- define "headmail.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "headmail.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "headmail.name" . -}}
{{- if hasPrefix $name .Release.Name }}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "headmail.labels" -}}
app.kubernetes.io/name: {{ include "headmail.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "headmail.serviceAccountName" -}}
{{- if .Values.serviceAccount.name }}
{{- .Values.serviceAccount.name -}}
{{- else -}}
{{- printf "%s-sa" (include "headmail.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/* Compute a sha256 checksum of the rendered config - used to trigger rollouts when config changes */}}
{{- define "headmail.configChecksum" -}}
{{- include (print .Template.BasePath "/config-configmap.yaml") . | sha256sum -}}
{{- end -}}

{{- define "headmail.secret.name" }}
{{- default (printf "%s-secret" (include "headmail.fullname" .)) .Values.secretEnv.secretName -}}
{{- end }}

{{- define "headmail.image" }}
{{- $registry := default .context.Values.image.registry .image.registry }}
{{- $tag := default .context.Chart.AppVersion .image.tag }}
{{- printf "%s/%s:%s" $registry .image.repository $tag -}}
{{- end }}
