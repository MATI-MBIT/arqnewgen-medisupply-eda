{{/*
Expand the name of the chart.
*/}}
{{- define "microservice.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "microservice.fullname" -}}
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
{{- define "microservice.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "microservice.labels" -}}
helm.sh/chart: {{ include "microservice.chart" . }}
{{ include "microservice.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "microservice.selectorLabels" -}}
app.kubernetes.io/name: {{ include "microservice.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "microservice.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "microservice.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate environment variables for MQTT configuration
*/}}
{{- define "microservice.mqttEnvVars" -}}
{{- if .Values.mqtt.enabled -}}
- name: MQTT_BROKER
  value: {{ .Values.mqtt.broker | quote }}
- name: MQTT_CLIENT_ID
  value: {{ printf "%s-%s" .Values.mqtt.clientId .Release.Name | quote }}
- name: MQTT_TOPIC
  value: {{ .Values.mqtt.topic | quote }}
- name: MQTT_USERNAME
  value: {{ .Values.mqtt.username | quote }}
- name: MQTT_PASSWORD
  value: {{ .Values.mqtt.password | quote }}
{{- end -}}
{{- end }}

{{/*
Generate environment variables from app.config
*/}}
{{- define "microservice.appConfigEnvVars" -}}
{{- range $key, $value := .Values.app.config -}}
- name: {{ $key }}
  value: {{ $value | quote }}
{{- end -}}
{{- end }}

{{/*
Generate all environment variables
*/}}
{{- define "microservice.envVars" -}}
{{- include "microservice.mqttEnvVars" . }}
{{- include "microservice.appConfigEnvVars" . }}
{{- with .Values.env.common }}
{{- toYaml . }}
{{- end }}
{{- with .Values.env.custom }}
{{- toYaml . }}
{{- end }}
{{- end }}