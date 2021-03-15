import { AuthState } from "../contexts/AuthContext";
import { LibraryCheckerServiceClient } from "./Library_checkerServiceClientPb";

const api_url = process.env.REACT_APP_API_URL;

export const authMetadata = (
  state: AuthState
):
  | {
      authorization: string;
    }
  | undefined => {
  if (!state.token) {
    return undefined;
  } else {
    return {
      authorization: "bearer " + state.token,
    };
  }
};

export default new LibraryCheckerServiceClient(
  api_url ?? "https://grpcweb-apiv1.yosupo.jp:443"
);
