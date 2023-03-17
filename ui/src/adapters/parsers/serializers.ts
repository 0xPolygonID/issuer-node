import dayjs from "dayjs";

import { CredentialAttribute, CredentialIssuePayload } from "src/adapters/api/credentials";
import {
  BooleanCredentialFormAttribute,
  CredentialForm,
  CredentialFormAttribute,
  DateCredentialFormAttribute,
  NumberCredentialFormAttribute,
  SingleChoiceCredentialFormAttribute,
} from "src/domain";

export const serializeBooleanCredentialFormAttribute = (
  booleanCredentialFormAttribute: BooleanCredentialFormAttribute
): CredentialAttribute => ({
  attributeKey: booleanCredentialFormAttribute.name,
  attributeValue: booleanCredentialFormAttribute.value ? 1 : 0,
});

export const serializeDateCredentialFormAttribute = (
  dateCredentialFormAttribute: DateCredentialFormAttribute
): CredentialAttribute => {
  const momentInstance = dayjs(dateCredentialFormAttribute.value);
  const numericDateString = momentInstance.format("YYYYMMDD");

  return {
    attributeKey: dateCredentialFormAttribute.name,
    attributeValue: parseInt(numericDateString),
  };
};

export const serializeNumberCredentialFormAttribute = (
  numberCredentialFormAttribute: NumberCredentialFormAttribute
): CredentialAttribute => ({
  attributeKey: numberCredentialFormAttribute.name,
  attributeValue: numberCredentialFormAttribute.value,
});

export const serializeSingleChoiceCredentialFormAttribute = (
  singleChoiceCredentialFormAttribute: SingleChoiceCredentialFormAttribute
): CredentialAttribute => ({
  attributeKey: singleChoiceCredentialFormAttribute.name,
  attributeValue: singleChoiceCredentialFormAttribute.value,
});

export const serializeCredentialFormAttribute = (
  credentialFormAttribute: CredentialFormAttribute
): CredentialAttribute => {
  switch (credentialFormAttribute.type) {
    case "boolean": {
      return serializeBooleanCredentialFormAttribute(credentialFormAttribute);
    }
    case "date": {
      return serializeDateCredentialFormAttribute(credentialFormAttribute);
    }
    case "number": {
      return serializeNumberCredentialFormAttribute(credentialFormAttribute);
    }
    case "singlechoice": {
      return serializeSingleChoiceCredentialFormAttribute(credentialFormAttribute);
    }
  }
};

export const serializeCredentialForm = (credentialForm: CredentialForm): CredentialIssuePayload => {
  const attributes = credentialForm.attributes.map(serializeCredentialFormAttribute);
  const expirationDate =
    credentialForm.expiration && dayjs(credentialForm.expiration).toISOString();
  const claimLinkExpiration =
    credentialForm.linkAccessibleUntil && credentialForm.linkAccessibleUntil.toISOString();
  const limitedClaims = credentialForm.linkMaximumIssuance;

  return {
    attributes,
    ...(expirationDate ? { expirationDate } : {}),
    ...(claimLinkExpiration ? { claimLinkExpiration } : {}),
    ...(limitedClaims ? { limitedClaims } : {}),
  };
};
