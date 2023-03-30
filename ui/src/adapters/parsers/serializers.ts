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

const serializeBooleanCredentialFormAttribute = (
  booleanCredentialFormAttribute: BooleanCredentialFormAttribute
): CredentialAttribute => ({
  attributeKey: booleanCredentialFormAttribute.name,
  attributeValue: booleanCredentialFormAttribute.value ? 1 : 0,
});

const serializeDateCredentialFormAttribute = (
  dateCredentialFormAttribute: DateCredentialFormAttribute
): CredentialAttribute => {
  const momentInstance = dayjs(dateCredentialFormAttribute.value);
  const numericDateString = momentInstance.format("YYYYMMDD");

  return {
    attributeKey: dateCredentialFormAttribute.name,
    attributeValue: parseInt(numericDateString),
  };
};

const serializeNumberCredentialFormAttribute = (
  numberCredentialFormAttribute: NumberCredentialFormAttribute
): CredentialAttribute => ({
  attributeKey: numberCredentialFormAttribute.name,
  attributeValue: numberCredentialFormAttribute.value,
});

const serializeSingleChoiceCredentialFormAttribute = (
  singleChoiceCredentialFormAttribute: SingleChoiceCredentialFormAttribute
): CredentialAttribute => ({
  attributeKey: singleChoiceCredentialFormAttribute.name,
  attributeValue: singleChoiceCredentialFormAttribute.value,
});

const serializeCredentialFormAttribute = (
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
  const expirationDate = credentialForm.expiration
    ? dayjs(credentialForm.expiration).toISOString()
    : null;
  const claimLinkExpiration = credentialForm.linkAccessibleUntil
    ? credentialForm.linkAccessibleUntil.toISOString()
    : null;
  const limitedClaims =
    credentialForm.linkMaximumIssuance !== undefined ? credentialForm.linkMaximumIssuance : null;

  return {
    attributes,
    claimLinkExpiration,
    expirationDate,
    limitedClaims,
  };
};
