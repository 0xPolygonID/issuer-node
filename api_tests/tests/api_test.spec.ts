import { test, expect } from '@playwright/test';

test.describe('API tests', () => {
    let did: string;
    test.beforeEach(async ({request}) => {
        const newDID = await request.post(`/v1/identities`, {
            data: {
                "didMetadata": {
                  "method": "iden3",
                  "blockchain": "privado",
                  "network": "main",
                  "type": "BJJ",
                  "authBJJCredentialStatus": "Iden3commRevocationStatusV1.0"
                }
              }
        });
        did = (await newDID.json()).identifier;
    });

    test('unsuported media type', async ({ request }) => {
        const newClaim = await request.post(`/v1/${did}/credentials`, {
            data: {
                "credentialSchema": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
                "type": "KYCAgeCredential",
                "credentialSubject": {
                    "id": "did:iden3:privado:main:2Sj3kThjNGrJN6Gk3QMaxtZrGC7DweusYtTWsr6jHm",
                    "birthday": 19960424,
                    "documentType": 2
                },
                "expiration": 1903357766
            }
        });
        expect(newClaim.ok()).toBeTruthy();

        var id = (await newClaim.json()).id;
        const agent = await request.post(`/v1/agent`, {
            data: {
                "id": "1924af5a-7d63-4850-addf-0177cdc34786",
                "thid": "1924af5a-7d63-4850-addf-0177cdc34786",
                "typ": "application/iden3comm-plain-json",
                "type": "https://iden3-communication.io/credentials/1.0/fetch-request",
                "body": {
                    "id": id,
                },
                "from": "did:iden3:privado:main:2ScShf8ab4s9kCnRzg6cf3Z81dTawxPLxjF5rEYgKv",
                "to": "did:iden3:privado:main:2SfrAF6Lya1HLEGWcSXTBMApk5YVmKR2ymZWDjZNSH"
            }
        });

        expect(agent.status()).toBe(400);
        expect((await agent.json()).message).toContain('unsupported media type');
    });

    test('revocation status request message type', async ({ request }) => {
        const newClaim = await request.post(`/v1/${did}/credentials`, {
            data: {
                "credentialSchema": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
                "type": "KYCAgeCredential",
                "credentialSubject": {
                    "id": "did:iden3:privado:main:2Sj3kThjNGrJN6Gk3QMaxtZrGC7DweusYtTWsr6jHm",
                    "birthday": 19960424,
                    "documentType": 2
                },
                "expiration": 1903357766
            }
        });
        expect(newClaim.ok()).toBeTruthy();

        var id = (await newClaim.json()).id;
        const claim = await request.get(`/v1/${did}/claims/${id}`);
        expect(claim.ok()).toBeTruthy();
        expect((await claim.json()).credentialStatus.type).toBe('Iden3commRevocationStatusV1.0');
        var claimData = await claim.json();
        const agent = await request.post(`/v1/agent`, {
            data: {
                "id": id,
                "thid": id,
                "type": "https://iden3-communication.io/revocation/1.0/request-status",
                "to": did,
                "from": claimData.credentialSubject.id,
                "typ": "application/iden3comm-plain-json",
                "body": {
                    "revocation_nonce": claimData.credentialStatus.revocationNonce,
                }
            }
        });
        expect(agent.ok()).toBeTruthy();
    });
});
