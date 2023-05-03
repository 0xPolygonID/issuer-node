import { AppError } from "src/domain";

interface RequestErrorResponse {
  error: AppError;
  success: false;
}

interface RequestSuccessResponse<D> {
  data: D;
  success: true;
}

export type RequestResponse<D> = RequestSuccessResponse<D> | RequestErrorResponse;
