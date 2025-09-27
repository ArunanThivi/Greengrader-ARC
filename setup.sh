
echo "🔧 GitHub Actions Runner Setup Wizard"
printf "Enter your GitHub org URL (e.g. https://github.com/my-org): "
read GITHUB_CONFIG_URL

printf "Create a new Application for your Github organization\n"
printf "Enter the newly made Application ID:\n"
read GITHUB_APP_ID

printf "Install the newly made application to your class organization\n"
printf "Enter the installation ID (seen in the URL after installing): \n"
read GITHUB_INSTALLATION_ID

printf "Download a Private Key from the App Settings Page\n"
printf "Enter the path to the private key: "
read GITHUB_PRIVATE_KEY_PATH

read -p "Minimum runners (Default 0): " MIN
read -p "Maximum runners (Default 0): " MAX


helm install arc \
    --namespace "greengrader-systems" \
    --create-namespace \
    oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set-controller

helm install "greengrader" \
    --namespace "greengrader-runners" \
    --create-namespace \
    --set githubConfigUrl="$GITHUB_CONFIG_URL" \
    --set githubConfigSecret="greengrader-auth" \
    ${MIN:+--set minRunners=$MIN} \
    ${MAX:+--set maxRunners=$MAX} \
    oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set

kubectl create secret generic greengrader-auth \
   --namespace=greengrader-runners \
   --from-literal=github_app_id=$GITHUB_APP_ID \
   --from-literal=github_app_installation_id=$GITHUB_INSTALLATION_ID \
   --from-file=github_app_private_key=$GITHUB_PRIVATE_KEY_PATH
