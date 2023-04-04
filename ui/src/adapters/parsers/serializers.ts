import dayjs from "dayjs";

import { CredentialAttribute, CredentialIssue } from "src/adapters/api/credentials";
import {
  BooleanCredentialFormAttribute,
  CredentialForm,
  CredentialFormAttribute,
  DateCredentialFormAttribute,
  NumberCredentialFormAttribute,
  SingleChoiceCredentialFormAttribute,
} from "src/domain/credential";

function serializeBooleanCredentialFormAttribute(
  booleanCredentialFormAttribute: BooleanCredentialFormAttribute
): CredentialAttribute {
  return {
    attributeKey: booleanCredentialFormAttribute.name,
    attributeValue: booleanCredentialFormAttribute.value ? 1 : 0,
  };
}

function serializeCredentialFormAttribute(
  credentialFormAttribute: CredentialFormAttribute
): CredentialAttribute {
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
}

function serializeDateCredentialFormAttribute(
  dateCredentialFormAttribute: DateCredentialFormAttribute
): CredentialAttribute {
  const momentInstance = dayjs(dateCredentialFormAttribute.value);
  const numericDateString = momentInstance.format("YYYYMMDD");

  return {
    attributeKey: dateCredentialFormAttribute.name,
    attributeValue: parseInt(numericDateString),
  };
}

function serializeNumberCredentialFormAttribute(
  numberCredentialFormAttribute: NumberCredentialFormAttribute
): CredentialAttribute {
  return {
    attributeKey: numberCredentialFormAttribute.name,
    attributeValue: numberCredentialFormAttribute.value,
  };
}

function serializeSingleChoiceCredentialFormAttribute(
  singleChoiceCredentialFormAttribute: SingleChoiceCredentialFormAttribute
): CredentialAttribute {
  return {
    attributeKey: singleChoiceCredentialFormAttribute.name,
    attributeValue: singleChoiceCredentialFormAttribute.value,
  };
}

// Exports

export function serializeCredentialForm(credentialForm: CredentialForm): CredentialIssue {
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
}
