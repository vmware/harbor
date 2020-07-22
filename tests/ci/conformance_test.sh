set -e

echo "get the conformance testing code..."
git clone https://github.com/opencontainers/distribution-spec.git

echo "create testing project"
STATUS=$(curl -w '%{http_code}' -H 'Content-Type: application/json' -H 'Accept: application/json' -X POST -u "admin:Harbor12345" -s --insecure "https://$IP/api/v2.0/projects" --data '{"project_name":"conformance","metadata":{"public":"false"},"storage_limit":-1}')
if [ $STATUS -ne 201 ]; then
		exit 1
fi

echo "run conformance test..."
export OCI_ROOT_URL="https://$1"
export OCI_NAMESPACE="conformance/testrepo"
export OCI_USERNAME="admin"
export OCI_PASSWORD="Harbor12345"
export OCI_DEBUG="true"

export OCI_TEST_PUSH=1
export OCI_TEST_PULL=1
export OCI_TEST_CONTENT_DISCOVERY=1
export OCI_TEST_CONTENT_MANAGEMENT=1

cd ./distribution-spec/conformance
go test .