import dayjs from "dayjs";

import { ClaimAttribute, ClaimIssuePayload } from "src/adapters/api/claims";
import {
  BooleanClaimFormAttribute,
  ClaimForm,
  ClaimFormAttribute,
  DateClaimFormAttribute,
  NumberClaimFormAttribute,
  SingleChoiceClaimFormAttribute,
} from "src/domain";

export const serializeBooleanClaimFormAttribute = (
  booleanClaimFormAttribute: BooleanClaimFormAttribute
): ClaimAttribute => ({
  attributeKey: booleanClaimFormAttribute.name,
  attributeValue: booleanClaimFormAttribute.value ? 1 : 0,
});

export const serializeDateClaimFormAttribute = (
  dateClaimFormAttribute: DateClaimFormAttribute
): ClaimAttribute => {
  const momentInstance = dayjs(dateClaimFormAttribute.value);
  const numericDateString = momentInstance.format("YYYYMMDD");

  return {
    attributeKey: dateClaimFormAttribute.name,
    attributeValue: parseInt(numericDateString),
  };
};

export const serializeNumberClaimFormAttribute = (
  numberClaimFormAttribute: NumberClaimFormAttribute
): ClaimAttribute => ({
  attributeKey: numberClaimFormAttribute.name,
  attributeValue: numberClaimFormAttribute.value,
});

export const serializeSingleChoiceClaimFormAttribute = (
  singleChoiceClaimFormAttribute: SingleChoiceClaimFormAttribute
): ClaimAttribute => ({
  attributeKey: singleChoiceClaimFormAttribute.name,
  attributeValue: singleChoiceClaimFormAttribute.value,
});

export const serializeClaimFormAttribute = (
  claimFormAttribute: ClaimFormAttribute
): ClaimAttribute => {
  switch (claimFormAttribute.type) {
    case "boolean": {
      return serializeBooleanClaimFormAttribute(claimFormAttribute);
    }
    case "date": {
      return serializeDateClaimFormAttribute(claimFormAttribute);
    }
    case "number": {
      return serializeNumberClaimFormAttribute(claimFormAttribute);
    }
    case "singlechoice": {
      return serializeSingleChoiceClaimFormAttribute(claimFormAttribute);
    }
  }
};

export const serializeClaimForm = (claimForm: ClaimForm): ClaimIssuePayload => {
  const attributes = claimForm.attributes.map(serializeClaimFormAttribute);
  const expirationDate = claimForm.expirationDate && dayjs(claimForm.expirationDate).toISOString();
  const claimLinkExpiration =
    claimForm.claimLinkExpiration && claimForm.claimLinkExpiration.toISOString();
  const limitedClaims = claimForm.limitedClaims;

  return {
    attributes,
    ...(expirationDate ? { expirationDate } : {}),
    ...(claimLinkExpiration ? { claimLinkExpiration } : {}),
    ...(limitedClaims ? { limitedClaims } : {}),
  };
};
