{{ define node_def }}
NODE_ADDRESS={{ .address | quote }}
NODE_PORT={{ .port | default "2376" }}
NODE_NAME={{ .name | default .address | quote }}
CONSUL={{ .consul | default "http://cf-consul:8500" | quote }}
NODE_CLUSTER={{ .cluster | "codefresh" }}
NODE_ROLE={{ .role | "builder" }}
{{ end }}