#!/bin/bash
#
# Script to trigger test_deploy_app.py.
# This is intended to be invoked as a step in Argo.
#
# test_delete.sh ${PROJECT} ${DEPLOYMENT}
set -ex

PROJECT=$1
DEPLOYMENT=$2
CLUSTER=$3
ZONE=$4
NAMESPACE=$5

gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}

gcloud container clusters get-credentials ${CLUSTER} --zone ${ZONE} --project ${PROJECT}
# delete test namespace
# kubectl delete namespace ${NAMESPACE}

for i in 1 2 3 4 5
do gcloud -q deployment-manager deployments delete ${DEPLOYMENT} --project=${PROJECT} && break || sleep 30
done

for i in 1 2 3 4 5
do gcloud -q source repos delete ${PROJECT}-kubeflow-config --force --project=${PROJECT} && break || sleep 30
done
