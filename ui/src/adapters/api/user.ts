import axios from "axios";
import { z } from "zod";
import { Response, buildErrorResponse, buildSuccessResponse } from "..";
import { datetimeParser, getStrictParser } from "../parsers";
import { buildAuthorizationHeader } from ".";
import { Env } from "src/domain";
import { Login, UserDetails, UserResponse, userProfile } from "src/domain/user";
import { API_VERSION } from "src/utils/constants";
import { List } from "src/utils/types";

export const userParser = getStrictParser<UserDetails, UserDetails>()(
  z.object({
    address: z.string(),
    adhar: z.string(),
    createdAt: datetimeParser,
    dob: z.string(),
    documentationSource: z.string(),
    gmail: z.string(),
    gstin: z.string(),
    id: z.string(),
    iscompleted: z.boolean(),
    name: z.string(),
    owner: z.string(),
    PAN: z.string(),
    phoneNumber: z.string(),
    username: z.string(),
    userType: z.string(),
  })
);

export const userResponseParser = getStrictParser<UserResponse, UserResponse>()(
  z.object({
    msg: z.string(),
    status: z.boolean(),
  })
);

export const loginParser = getStrictParser<Login, Login>()(
  z.object({
    fullName: z.string(),
    gmail: z.string(),
    iscompleted: z.boolean(),
    password: z.string(),
    userDID: z.string(),
    username: z.string(),
    userType: z.string(),
  })
);

export async function getUserDetails({
  env,
  password,
  username,
}: {
  env: Env;
  password: string;
  username: string;
}): Promise<Response<UserDetails>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: { password, username },
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/getUser`,
    });
    return buildSuccessResponse(userParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function login({
  env,
  password,
  username,
}: {
  env: Env;
  password: string;
  username: string;
}): Promise<Response<Login>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: { password, username },
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/login`,
    });

    return buildSuccessResponse(loginParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function updateUser({
  env,
  updatePayload,
}: {
  env: Env;
  updatePayload: userProfile;
}): Promise<Response<UserResponse>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      data: updatePayload,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      url: `${API_VERSION}/updateUser`,
    });

    return buildSuccessResponse(userResponseParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}

export async function getUser({
  env,
  signal,
  userDID,
}: {
  env: Env;
  signal?: AbortSignal;
  userDID: string;
}): Promise<Response<List<UserDetails>>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "POST",
      signal,
      url: `${API_VERSION}/getUser/${userDID}`,
    });

    return buildSuccessResponse(userParser.parse(response.data));
  } catch (error) {
    return buildErrorResponse(error);
  }
}
