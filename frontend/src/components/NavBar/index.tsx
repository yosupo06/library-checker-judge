import Button from "@mui/material/Button";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import Menu from "@mui/material/Menu";
import { GitHub } from "@mui/icons-material";
import MenuIcon from "@mui/icons-material/Menu";
import React from "react";
import { Link, LinkProps, useNavigate } from "react-router-dom";
import { useLangContext } from "../../contexts/LangContext";
import flagJA from "./flag_ja.svg";
import flagUS from "./flag_us.svg";
import { styled } from "@mui/system";
import { useCurrentUser } from "../../api/client_wrapper";
import { Drawer, IconButton } from "@mui/material";
import { useSignOutMutation } from "../../auth/auth";

const ButtonLink = styled(Button)<LinkProps>();

const NavBar: React.FC = () => {
  const title = (
    <>
      <ButtonLink
        LinkComponent={Link}
        color="inherit"
        to="/"
        sx={{
          fontSize: "16px",
          fontWeight: "bolder",
        }}
      >
        Library Checker
      </ButtonLink>
    </>
  );

  const [open, setOpen] = React.useState(false);

  const toggleDrawerOpen = () => {
    setOpen(!open);
  };

  return (
    <>
      <AppBar position="static">
        <Toolbar key="bar-md" sx={{ display: { xs: "none", md: "flex" } }}>
          {title}
          <NavBarElements />
        </Toolbar>

        <Toolbar key="bar-xs" sx={{ display: { xs: "flex", md: "none" } }}>
          {title}
          <IconButton
            onClick={toggleDrawerOpen}
            sx={{
              ml: "auto",
              mr: 0.2,
            }}
            color="inherit"
          >
            <MenuIcon />
          </IconButton>
        </Toolbar>
      </AppBar>
      <Drawer
        sx={{
          display: { xs: "flex", md: "none" },
        }}
        PaperProps={{
          sx: { width: "35%" },
        }}
        anchor="left"
        open={open}
        onClose={toggleDrawerOpen}
      >
        <NavBarElements />
      </Drawer>
    </>
  );
};

export default NavBar;

const NavBarElements: React.FC = () => {
  const user = useCurrentUser();

  const isDeveloper = user.isSuccess && user.data.user?.isDeveloper;

  return (
    <>
      {/* left side */}
      <ButtonLink LinkComponent={Link} color="inherit" to="/submissions">
        Submissions
      </ButtonLink>
      <ButtonLink LinkComponent={Link} color="inherit" to="/ranking">
        Ranking
      </ButtonLink>
      <ButtonLink LinkComponent={Link} color="inherit" to="/help">
        Help
      </ButtonLink>
      {isDeveloper && <ToolsMenu />}

      {/* right side */}
      <LangSelect />
      {UserMenu()}
      <Button
        color="inherit"
        href="https://github.com/yosupo06/library-checker-problems"
        target="_blank"
        rel="noopener"
      >
        <GitHub />
      </Button>
    </>
  );
};

const LangSelect = () => {
  const lang = useLangContext();

  return (
    <Select
      value={lang?.state.lang}
      variant="outlined"
      onChange={(e) =>
        lang?.dispatch({
          type: "change",
          payload: (e.target.value as string) === "ja" ? "ja" : "en",
        })
      }
      sx={{
        ml: "auto",
        mr: { xs: "auto", md: 0.2 },
      }}
    >
      <MenuItem value="en">
        <img src={flagUS} alt="us" height="15px"></img>
      </MenuItem>
      <MenuItem value="ja">
        <img src={flagJA} alt="ja" height="15px"></img>
      </MenuItem>
    </Select>
  );
};

const NavbarLink = styled(Link)({
  color: "inherit",
  textDecoration: "none",
  textTransform: "none",
});

const NavbarButton = styled(Button)({
  textDecoration: "none",
  textTransform: "none",
});

const ToolsMenu: React.FC = () => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

  const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <>
      <NavbarButton color="inherit" onClick={handleClick}>
        Tools
      </NavbarButton>
      <Menu
        anchorEl={anchorEl}
        anchorOrigin={{
          vertical: "bottom",
          horizontal: "center",
        }}
        transformOrigin={{
          vertical: "top",
          horizontal: "center",
        }}
        open={Boolean(anchorEl)}
        onClose={handleClose}
      >
        <MenuItem>
          <NavbarLink to={`/tool/statementviewer`}>Statement Viewer</NavbarLink>
        </MenuItem>
      </Menu>
    </>
  );
};

const UserMenu = () => {
  const navigate = useNavigate();

  const signOutMutation = useSignOutMutation();
  const currentUser = useCurrentUser();

  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

  const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    signOutMutation.mutate();
    handleClose();
  };

  if (!currentUser.isSuccess || currentUser.data.user == null) {
    return [
      <NavbarButton color="inherit" onClick={() => navigate("/register")}>
        Register
      </NavbarButton>,
      <NavbarButton color="inherit" onClick={() => navigate("/login")}>
        Login
      </NavbarButton>,
    ];
  }

  return [
    <NavbarButton color="inherit" onClick={handleClick}>
      {currentUser.data.user.name}
    </NavbarButton>,
    <Menu
      anchorEl={anchorEl}
      anchorOrigin={{
        vertical: "bottom",
        horizontal: "center",
      }}
      transformOrigin={{
        vertical: "top",
        horizontal: "center",
      }}
      open={Boolean(anchorEl)}
      onClose={handleClose}
    >
      <MenuItem>
        <NavbarLink to={`/profile`}>Profile</NavbarLink>
      </MenuItem>
      <MenuItem onClick={handleLogout}>Logout</MenuItem>
    </Menu>,
  ];
};
