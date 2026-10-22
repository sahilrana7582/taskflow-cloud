
platform-gitops/
├── README.md
│
├── bootstrap/
│   └── argocd/
│       ├── namespace.yaml
│       └── root-application.yaml
│
├── applications/
│   ├── platform/
│   │   ├── cert-manager.yaml
│   │   ├── external-secrets.yaml
│   │   ├── ingress-controller.yaml
│   │   └── monitoring.yaml
│   │
│   ├── data/
│   │   ├── postgresql.yaml
│   │   ├── mongodb.yaml
│   │   ├── redis.yaml
│   │   └── kafka.yaml
│   │
│   └── services/
│       ├── user-service.yaml
│       ├── order-service.yaml
│       ├── payment-service.yaml
│       └── notification-service.yaml
│
├── infrastructure/
│   ├── databases/
│   │   ├── postgresql/
│   │   │   ├── base/
│   │   │   │   ├── kustomization.yaml
│   │   │   │   ├── namespace.yaml
│   │   │   │   ├── service.yaml
│   │   │   │   └── network-policy.yaml
│   │   │   └── overlays/
│   │   │       ├── dev/
│   │   │       ├── staging/
│   │   │       └── prod/
│   │   │
│   │   ├── mongodb/
│   │   │   ├── base/
│   │   │   └── overlays/
│   │   │       ├── dev/
│   │   │       ├── staging/
│   │   │       └── prod/
│   │   │
│   │   └── redis/
│   │       ├── base/
│   │       └── overlays/
│   │
│   └── messaging/
│       └── kafka/
│           ├── base/
│           └── overlays/
│
├── services/
│   ├── user-service/
│   │   ├── base/
│   │   │   ├── kustomization.yaml
│   │   │   ├── deployment.yaml
│   │   │   ├── service.yaml
│   │   │   ├── service-account.yaml
│   │   │   ├── configmap.yaml
│   │   │   ├── network-policy.yaml
│   │   │   └── pdb.yaml
│   │   └── overlays/
│   │       ├── dev/
│   │       ├── staging/
│   │       └── prod/
│   │
│   ├── order-service/
│   │   ├── base/
│   │   └── overlays/
│   │
│   ├── payment-service/
│   │   ├── base/
│   │   └── overlays/
│   │
│   └── notification-service/
│       ├── base/
│       └── overlays/
│
├── environments/
│   ├── dev/
│   │   └── kustomization.yaml
│   ├── staging/
│   │   └── kustomization.yaml
│   └── prod/
│       └── kustomization.yaml
│
├── policies/
│   ├── network/
│   ├── resource-quotas/
│   └── pod-security/
│
└── observability/
    ├── dashboards/
    ├── alerts/
    └── logging/
