import Button, { ButtonProps } from "@mui/material/Button";
import React from "react";
import { Link, To } from "react-router-dom";

interface LinkButtonProps extends ButtonProps {
  to?: To;
}
export const LinkButton: React.FC<LinkButtonProps> = (props) => (
  <Button LinkComponent={Link} variant="outlined" {...props} />
);

interface ExternalLinkButtonProps {
  startIcon: ButtonProps["startIcon"];
  href: string;
  children?: React.ReactNode;
}
export const ExternalLinkButton: React.FC<ExternalLinkButtonProps> = (
  props,
) => (
  <Button
    variant="outlined"
    target="_blank"
    rel="noopener noreferrer"
    {...props}
  />
);
