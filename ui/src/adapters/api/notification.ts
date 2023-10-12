import axios from "axios";
import { z } from "zod";

import { Response, buildErrorResponse, buildSuccessResponse } from "src/adapters";
import { buildAuthorizationHeader } from "src/adapters/api";
import { datetimeParser, getListParser, getStrictParser } from "src/adapters/parsers";
import { Env, Notification } from "src/domain";

import { API_VERSION } from "src/utils/constants";
import { List } from "src/utils/types";

type NotificationInput = Omit<Notification, "proofTypes" | "createdAt" | "expiresAt"> & {
  createdAt: string;
  expiresAt: string | null;
};

export type NotificationStatus = "all" | "revoked" | "expired";
export const notificationStatusParser = getStrictParser<NotificationStatus>()(
  z.union([z.literal("all"), z.literal("revoked"), z.literal("expired")])
);

export const NotificationParser = getStrictParser<NotificationInput, Notification>()(
  z.object({
    created_at: datetimeParser,
    id: z.string(),
    module: z.string(),
    notification_message: z.string(),
    notification_title: z.string(),
    notification_type: z.string(),
    user_id: z.string(),
  })
);
export async function getNotification({
  env,
  module,
  signal,
}: {
  env: Env;
  module: string;
  params?: {
    query?: string;
  };
  signal?: AbortSignal;
}): Promise<Response<List<Notification>>> {
  try {
    const response = await axios({
      baseURL: env.api.url,
      headers: {
        Authorization: buildAuthorizationHeader(env),
      },
      method: "GET",
      signal,
      url: `${API_VERSION}/notifications/${module}`,
    });
    return buildSuccessResponse(
      getListParser(NotificationParser)
        .transform(({ failed, successful }) => ({
          failed,
          successful: successful.sort((a, b) => b.created_at.getTime() - a.created_at.getTime()),
        }))
        .parse(response.data)
    );
  } catch (error) {
    return buildErrorResponse(error);
  }
}
