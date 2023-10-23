import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Paper from "@mui/material/Paper";
import TablePagination from "@mui/material/TablePagination";
import FormControl from "@mui/material/FormControl";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import ListSubheader from "@mui/material/ListSubheader";
import React, { useState } from "react";
import { useLocation } from "react-use";
import client, {
    useLangList,
    useProblemCategories,
    useProblemList,
    useSubmissionList,
} from "../api/client_wrapper";
import SubmissionTable from "../components/SubmissionTable";
import { categoriseProblems } from "../utils/ProblemCategorizer";
import { styled } from "@mui/system";
import KatexTypography from "../components/katex/KatexTypography";
import { Container } from "@mui/material";
import { getApp, initializeApp } from "firebase/app";
import { connectAuthEmulator, createUserWithEmailAndPassword, getAuth, getIdToken, sendPasswordResetEmail, signInWithEmailAndPassword, signInWithRedirect, signOut } from "firebase/auth";
import { GoogleAuthProvider } from "firebase/auth/cordova";
import { QueryClient, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CurrentUserInfoRequest } from "../proto/library_checker";

const config = {
    apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
    authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
};

const app = initializeApp(config);
const auth = getAuth(app)
const emulatorUrl = import.meta.env.VITE_FIREBASE_AUTH_EMULATOR_URL
if (emulatorUrl && !auth.emulatorConfig) {
    connectAuthEmulator(auth, emulatorUrl)
}

export const useAuth = () => {    
    return auth
}

const currentUserQueryKey = ["auth", "currentUser"]

export const registerQueryClient = (queryClient: QueryClient) => {
    auth.onAuthStateChanged((user) => {
        if (user !== queryClient.getQueryData(currentUserQueryKey)) {
            queryClient.setQueryData(currentUserQueryKey, user)
        }
    })
}

export const useCurrentAuthUser = () => {
    const queryClient = useQueryClient()
    const auth = useAuth()
    return useQuery(
        currentUserQueryKey,
        () => {
            return auth.currentUser
        }
    );
}

export const useIdToken = () => {
    const currentAuthUser = useCurrentAuthUser()
    return useQuery(
        ["auth", "idToken", currentAuthUser.data?.email],
        () => {
            return currentAuthUser.data?.getIdToken() ?? Promise.resolve(null)
        },
    ) 
}

export const useSignInMutation = () => {
    const auth = useAuth()

    return useMutation((args: {email: string, password: string}) => {
        return signInWithEmailAndPassword(auth, args.email, args.password)
    })
}

export const useSignOutMutation = () => {
    const auth = useAuth()

    return useMutation(() => {
        return signOut(auth)
    })
}
