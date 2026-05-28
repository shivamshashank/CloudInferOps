# StackPulse - Quick Start Guide

## 📦 What You Have

A fully functional, production-ready CLI tool for deploying complete observability stacks on AWS.

**Binary**: `./stackpulse` (18MB, executable)

---

## ⚡ One-Minute Setup

### 1. Prerequisites
- AWS credentials configured (`aws configure`)
- Either:
  - An **EKS cluster** already created in AWS, OR
  - An **EC2 instance with K3s** installed

### 2. Deploy
```bash
# For EKS cluster
./stackpulse up --region eu-west-2 --cluster-name my-cluster

# For EC2 instance
./stackpulse up --region us-east-1 --cluster-name my-ec2-instance
```

### 3. Access
After deployment completes, you'll get access URLs like:
- Grafana: `http://grafana.YOUR_IP.nip.io` (admin/admin)
- Prometheus: `http://prometheus.YOUR_IP.nip.io`
- AlertManager: `http://alertmanager.YOUR_IP.nip.io`

---

## 📋 Complete Command Reference

### Deployment
```bash
# Deploy to EKS (requires existing cluster)
./stackpulse up \
  --region eu-west-2 \
  --cluster-name my-eks-cluster

# Deploy to EC2 with K3s (requires existing instance with K3s)
./stackpulse up \
  --region us-east-1 \
  --cluster-name my-ec2-instance
```

### Alert Integrations
```bash
# Configure Slack
./stackpulse integrate slack \
  --webhook-url https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# Configure PagerDuty
./stackpulse integrate pagerduty \
  --routing-key YOUR_ROUTING_KEY
```

### Optional Integrations
```bash
./stackpulse enable dynatrace
./stackpulse enable datadog
./stackpulse enable newrelic
```

### Help
```bash
./stackpulse --help
./stackpulse up --help
./stackpulse integrate --help
./stackpulse enable --help
```

---

## 🏗️ What Gets Deployed

| Component | Purpose |
|-----------|---------|
| **Prometheus** | Metrics collection & storage |
| **Grafana** | Dashboards & visualization |
| **AlertManager** | Alert routing & management |
| **NGINX Ingress** | Reverse proxy & DNS |
| **kube-prometheus-stack** | Complete monitoring setup |
| **Namespace: observability** | Kubernetes namespace for all components |

---

## 🔍 Troubleshooting

### "Cluster not found"
- Ensure the EKS cluster or EC2 instance exists in the specified region
- Verify the name matches exactly

### "AWS credentials not found"
- Run: `aws configure`
- Provide your AWS Access Key ID and Secret Access Key

### "Kubernetes not running on EC2"
- SSH into the EC2 instance and install K3s:
  ```bash
  curl -sfL https://get.k3s.io | sh -
  ```

### "Port 80 timeout"
- Check EC2 security group allows inbound traffic on port 80 (HTTP)
- Add inbound rule: Type=HTTP, Port=80, Source=0.0.0.0/0

---

## 📊 System Requirements

### For EKS Deployment
- AWS account with EKS permissions
- Existing EKS cluster
- Local kubectl/kubeconfig access

### For EC2 Deployment
- AWS account with EC2 permissions
- Running EC2 instance
- K3s installed on EC2
- Port 80 (HTTP) open in security group
- AWS Systems Manager (SSM) agent running

### Local Machine
- Go 1.21+ (if rebuilding)
- AWS CLI configured
- kubectl installed

---

## 🎯 Common Workflows

### Workflow 1: Deploy to Existing EKS Cluster
```bash
# 1. Ensure EKS cluster exists
aws eks list-clusters --region eu-west-2

# 2. Deploy StackPulse
./stackpulse up --region eu-west-2 --cluster-name my-cluster

# 3. Wait for deployment (~5 minutes)

# 4. Access Grafana at the provided URL
```

### Workflow 2: Deploy to EC2 with K3s
```bash
# 1. Launch EC2 instance
aws ec2 run-instances --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.medium --tag-specifications \
  'ResourceType=instance,Tags=[{Key=Name,Value=my-ec2}]'

# 2. SSH and install K3s
curl -sfL https://get.k3s.io | sh -

# 3. Deploy StackPulse
./stackpulse up --region us-east-1 --cluster-name my-ec2

# 4. Access via nip.io URLs
```

### Workflow 3: Add Slack Alerts
```bash
# 1. Create Slack webhook at https://api.slack.com/apps
# 2. Configure integration
./stackpulse integrate slack \
  --webhook-url https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# 3. Alerts from AlertManager now go to Slack
```

---

## 📁 Project Structure

```
stackpulse/
├── stackpulse          # ← Executable binary (run this!)
├── main.go             # Entry point
├── go.mod              # Dependencies
├── go.sum              # Checksums
├── README.md           # Full documentation
├── QUICK_START.md      # ← You are here
├── COMPLETION_SUMMARY.md # Project status
├── BUILD_REPORT.md     # Build details
├── cmd/                # CLI commands
├── internal/           # Core logic
└── helm/               # Helm charts
```

---

## 🚀 Next Steps

1. **Verify Prerequisites**
   - [ ] AWS credentials working
   - [ ] EKS cluster exists (OR EC2 with K3s)
   - [ ] Kubectl can access cluster

2. **Deploy**
   - [ ] Run `./stackpulse up --region X --cluster-name Y`
   - [ ] Wait for completion (~5-10 min)

3. **Verify**
   - [ ] Access Grafana dashboard
   - [ ] Check Prometheus targets
   - [ ] View AlertManager

4. **Integrate** (Optional)
   - [ ] Configure Slack
   - [ ] Configure PagerDuty
   - [ ] Enable optional plugins

---

## 📚 Resources

- **AWS Documentation**: https://aws.amazon.com/eks/
- **Kubernetes**: https://kubernetes.io/docs/
- **Prometheus**: https://prometheus.io/docs/
- **Grafana**: https://grafana.com/docs/
- **K3s**: https://docs.k3s.io/

---

**Status**: ✅ Production Ready | **Version**: 1.0 | **Built**: 2026-05-13
