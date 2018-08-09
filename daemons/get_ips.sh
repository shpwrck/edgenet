#!/bin/bash
DELIMETER='{{"\n"}}'
FIELD='Hostname'
TEMPLATE="{{range.items}}{{range.status.addresses}}{{if eq .type \"$FIELD\"}}{'name': '{{.address}}'},{{end}}{{end}}{{.status.conditions}}$DELIMETER{{end}}"
kubectl get nodes -o template --template="${TEMPLATE}"