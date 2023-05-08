import { AppError } from "src/domain";
import { buildAppError } from "src/utils/error";

interface SuccessResponse<D> {
  data: D;
  success: true;
}
interface ErrorResponse {
  error: AppError;
  success: false;
}

export type Response<D> = SuccessResponse<D> | ErrorResponse;

export function buildSuccessResponse<D>(data: D): SuccessResponse<D> {
  return { data, success: true };
}

export function buildErrorResponse(error: unknown): ErrorResponse {
  return { error: buildAppError(error), success: false };
}
