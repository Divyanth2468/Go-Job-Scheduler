import json
import requests

def lambda_handler(event, context):
    url = event.get("url", "https://example.com")
    try:
        response = requests.get(url)
        return {
            "statusCode": response.status_code,
            "body": response.text[:200]
        }
    except Exception as e:
        return {
            "statusCode": 500,
            "error": str(e)
        }
