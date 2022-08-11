from httplib2 import Http
from googleapiclient import discovery
from oauth2client.client import GoogleCredentials
import json
from os import getenv
from requests import post
from google.cloud import secretmanager

TARGET_ZONES = ['asia-northeast1-a', 'asia-northeast1-b', 'asia-northeast1-c',
                'asia-east1-a', 'asia-east1-b', 'asia-east1-c']

def main_handler(event, context):
    hook_url = getenv("HOOK_URL")

    credentials = GoogleCredentials.get_application_default()

    service = discovery.build(
        'compute', 'v1',
        http=credentials.authorize(Http()),
        cache_discovery=False
    )

    fields = []
    for zone in TARGET_ZONES:
        instances = service.instances().list(
            project='library-checker-project', zone=zone).execute()
        for instance in instances.get('items', []):
            title = instance['status']
            value = ""
            if instance['scheduling']['preemptible']:
                value = "{}({} PREEMPTIBLE)"
            else:
                value = "{}({})"
            value = value.format(
                instance['name'], instance['machineType'].split('/')[-1])
            fields.append({
                "name": title,
                "value": value,
            })

    payload = {
        "embeds": [{
            "title": "Judge server status",
            "fields": fields
        }]
    }

    result = post(hook_url, data=json.dumps(payload), headers={
                  'content-type': 'application/json'})
    print(result)


def main():
    main_handler(None, None)


if __name__ == "__main__":
    main()
