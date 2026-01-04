{{/*
Generate the name of a Secret.

Usage:
{{ include "common.secrets.name" (dict "existingSecret" .Values.path.to.existingSecret "defaultNameSuffix" "my-suffix" "context" .) }}

Params:
- existingSecret - Optional. The generated name will be based on this if provided.
- defaultNameSuffix - Optional. Appended as suffix if default name is used.
- context - Required. The context for the template evaluation.
*/}}
{{- define "common.secrets.name" -}}
{{- $name := printf "%s-%s" (include "common.names.fullname" .context) (default "" .defaultNameSuffix) | trunc 63 | trimSuffix "-" }}
{{- if (.existingSecret).name -}}
{{- $name = tpl .existingSecret.name .context -}}
{{- end -}}
{{- printf "%s" $name -}}
{{- end -}}

{{/*
Generate the secret key.

Usage:
{{ include "common.secrets.key" (dict "existingSecret" .Values.path.to.existingSecret "key" "keyName" "context" .) }}

Params:
- existingSecret - Optional. The path to existing secrets config.
- key - Required. Name of the key in the secret.
- context - Required. The context for the template evaluation.
*/}}
{{- define "common.secrets.key" -}}
{{- $_ := required "Variable .key is required" .key -}}
{{- $_ := required "Variable .context is required" .context -}}
{{- $customKey := get ( default dict (.existingSecret).keyMapping ) .key -}}
{{- if $customKey -}}
{{- tpl $customKey .context -}}
{{- else -}}
{{- .key -}}
{{- end -}}
{{- end -}}
