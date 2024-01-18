import { initializeApp } from "firebase/app";
import {
  connectAuthEmulator,
  createUserWithEmailAndPassword,
  getAuth,
  sendPasswordResetEmail,
  signInWithEmailAndPassword,
  signOut,
  sendEmailVerification,
  updateEmail,
  verifyBeforeUpdateEmail,
} from "firebase/auth";
import { QueryClient, useMutation, useQuery } from "@tanstack/react-query";

const config = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
};

const app = initializeApp(config);
const auth = getAuth(app);
const emulatorUrl = import.meta.env.VITE_FIREBASE_AUTH_EMULATOR_URL;
if (emulatorUrl && !auth.emulatorConfig) {
  connectAuthEmulator(auth, emulatorUrl);
}

const currentUserQueryKey = ["auth", "currentUser"];

export const registerQueryClient = (queryClient: QueryClient) => {
  auth.onAuthStateChanged((user) => {
    if (user !== queryClient.getQueryData(currentUserQueryKey)) {
      queryClient.setQueryData(currentUserQueryKey, user);
    }
  });
};

export const useCurrentAuthUser = () => {
  return useQuery(currentUserQueryKey, () =>
    auth.authStateReady().then(() => auth.currentUser)
  );
};

export const useIdToken = () => {
  const currentAuthUser = useCurrentAuthUser();
  return useQuery({
    queryKey: ["auth", "idToken", currentAuthUser.data?.email],
    queryFn: () => currentAuthUser.data?.getIdToken() ?? Promise.resolve(null),
    enabled: !currentAuthUser.isLoading,
  });
};

export const useRegisterMutation = () => {
  return useMutation((args: { email: string; password: string }) => {
    return createUserWithEmailAndPassword(
      auth,
      args.email,
      args.password
    ).catch((error) => {
      if (error.code === "auth/email-already-in-use") {
        return signInWithEmailAndPassword(auth, args.email, args.password);
      } else {
        throw error;
      }
    });
  });
};

export const useSignInMutation = () => {
  return useMutation((args: { email: string; password: string }) => {
    return signInWithEmailAndPassword(auth, args.email, args.password);
  });
};

export const useSignOutMutation = () => {
  return useMutation(() => {
    return signOut(auth);
  });
};

export const useUpdateEmailMutation = () => {
  return useMutation(async (newEmail: string) => {
    const user = auth.currentUser;
    if (!user) {
      return Promise.reject();
    }
    await verifyBeforeUpdateEmail(auth.currentUser, newEmail);
  });
};

export const useSendPasswordResetEmailMutation = () => {
  return useMutation((email: string) => {
    return sendPasswordResetEmail(auth, email);
  });
};
