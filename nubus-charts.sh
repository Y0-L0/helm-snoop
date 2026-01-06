charts=(
  "oci://artifacts.software-univention.de/nubus/charts/nubus-common"
  "oci://artifacts.software-univention.de/nubus/charts/guardian"
  "oci://artifacts.software-univention.de/nubus/charts/license-import"
  "oci://artifacts.software-univention.de/nubus/charts/ldap-notifier"
  "oci://artifacts.software-univention.de/nubus/charts/ldap-server"
  "oci://artifacts.software-univention.de/nubus/charts/notifications-api"
  "oci://artifacts.software-univention.de/nubus/charts/portal-frontend"
  "oci://artifacts.software-univention.de/nubus/charts/portal-consumer"
  "oci://artifacts.software-univention.de/nubus/charts/portal-server"
  "oci://artifacts.software-univention.de/nubus/charts/selfservice-consumer"
  "oci://artifacts.software-univention.de/nubus/charts/provisioning"
  "oci://artifacts.software-univention.de/nubus/charts/udm-listener"
  "oci://artifacts.software-univention.de/nubus/charts/keycloak"
  "oci://artifacts.software-univention.de/nubus/charts/keycloak-extensions"
  "oci://artifacts.software-univention.de/nubus/charts/keycloak-bootstrap"
  "oci://artifacts.software-univention.de/nubus/charts/stack-data-ums"
  "oci://artifacts.software-univention.de/nubus/charts/twofa-helpdesk"
  "oci://artifacts.software-univention.de/nubus/charts/udm-rest-api"
  "oci://artifacts.software-univention.de/nubus/charts/umc-gateway"
  "oci://artifacts.software-univention.de/nubus/charts/umc-server"
  "oci://artifacts.software-univention.de/nubus/charts/scim-server"
)

for chart_url in "${charts[@]}"; do
  chart_name=$(basename $chart_url)
  echo -e "\n-------------- Chart: $chart_name --------------\n"
  # helm pull $chart_url
  ./helm-snoop $chart_name*
done
