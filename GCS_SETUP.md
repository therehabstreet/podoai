// GCS Signed URL Implementation Guide

## Environment Variables

Set these environment variables to configure GCS signed URL generation:

```bash
# Required: GCS bucket name
export GCS_BUCKET_NAME="podoai-scans"

# Optional: Service account key file path (for local development)
export GCS_SERVICE_ACCOUNT_KEY_PATH="/path/to/service-account-key.json"

# Optional: Signed URL expiry time in minutes (default: 15)
export GCS_SIGNED_URL_EXPIRY_MINUTES=15

# Alternative: Use Application Default Credentials
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

## Authentication Methods

The implementation supports multiple authentication methods in order of preference:

1. **Service Account Key File**: Specified via `GCS_SERVICE_ACCOUNT_KEY_PATH`
2. **Application Default Credentials**: Via `GOOGLE_APPLICATION_CREDENTIALS` 
3. **GCP Metadata Service**: When running on GCP (Compute Engine, Cloud Run, etc.)
4. **Fallback**: Placeholder URLs for development when no credentials are available

## Production Setup

### 1. Create Service Account
```bash
# Create service account
gcloud iam service-accounts create podoai-media-signer \
    --display-name="PodoAI Media URL Signer"

# Grant Storage Object Viewer permission (for signed URLs)
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
    --member="serviceAccount:podoai-media-signer@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/storage.objectViewer"

# Create and download key
gcloud iam service-accounts keys create podoai-service-account.json \
    --iam-account=podoai-media-signer@YOUR_PROJECT_ID.iam.gserviceaccount.com
```

### 2. Deploy with Service Account
```bash
# Set environment variable
export GCS_SERVICE_ACCOUNT_KEY_PATH="/path/to/podoai-service-account.json"

# Or use Application Default Credentials
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/podoai-service-account.json"
```

### 3. Cloud Run Deployment
When deploying to Cloud Run, you can attach the service account directly:

```bash
gcloud run deploy podoai-api \
    --image=gcr.io/YOUR_PROJECT/podoai \
    --service-account=podoai-media-signer@YOUR_PROJECT_ID.iam.gserviceaccount.com \
    --set-env-vars="GCS_BUCKET_NAME=podoai-scans"
```

## Development Setup

For local development without GCS credentials:

```bash
# The API will automatically fallback to placeholder URLs
export GCS_BUCKET_NAME="podoai-scans-dev"
# No credentials needed - will use placeholders
```

## API Response

The API now generates real signed URLs when properly configured:

```json
{
  "images": [
    {
      "type": "LEFT_FOOT_DORSAL",
      "signed_url": "https://storage.googleapis.com/podoai-scans/scans/clinic_456/scan_123/media/LEFT_FOOT_DORSAL.jpg?X-Goog-Algorithm=GOOG4-RSA-SHA256&X-Goog-Credential=...",
      "thumbnail_url": "https://storage.googleapis.com/podoai-scans/scans/clinic_456/scan_123/media/thumbnails/LEFT_FOOT_DORSAL.jpg?X-Goog-Algorithm=GOOG4-RSA-SHA256&X-Goog-Credential=...",
      "gcs_path": "/scans/clinic_456/scan_123/media/LEFT_FOOT_DORSAL.jpg",
      "thumbnail_path": "/scans/clinic_456/scan_123/media/thumbnails/LEFT_FOOT_DORSAL.jpg",
      "expires_at": "2025-10-12T15:30:00Z"
    }
  ]
}
```

## Testing

You can test the implementation by:

1. **Without Credentials**: Will return placeholder URLs
2. **With Service Account**: Will return real signed URLs
3. **Invalid Bucket**: Will fallback to placeholder URLs with warning

The fallback mechanism ensures the API continues to work in all environments while providing real functionality when properly configured.