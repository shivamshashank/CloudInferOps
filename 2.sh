# Make the binary executable
chmod +x ./stackpulse

# 1. Run pre-flight checks
./stackpulse doctor

# 2. Deploy observability stack (uses sudo to access root kubeconfig)
sudo ./stackpulse deploy observability

# 3. Retrieve plaintext credentials and access links
sudo ./stackpulse connect --browser=false