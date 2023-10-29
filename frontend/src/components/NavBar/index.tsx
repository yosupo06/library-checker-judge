import Button from "@mui/material/Button";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import Menu from "@mui/material/Menu";
import { GitHub } from "@mui/icons-material";
import MenuIcon from "@mui/icons-material/Menu";
import React, { useContext } from "react";
import { Link, useNavigate } from "react-router-dom";
import { AuthContext } from "../../contexts/AuthContext";
import { LangContext } from "../../contexts/LangContext";
import flagJA from "./flag_ja.svg";
import flagUS from "./flag_us.svg";
import { styled } from "@mui/system";
import { useCurrentUser, useUserInfo } from "../../api/client_wrapper";
import { Drawer, IconButton } from "@mui/material";
import { useSignOutMutation } from "../../auth/auth";

const NavbarLink = styled(Link)({
  color: "inherit",
  textDecoration: "none",
  textTransform: "none",
});

const NavbarButton = styled(Button)({
  textDecoration: "none",
  textTransform: "none",
});

const LangSelect = () => {
  const lang = useContext(LangContext);

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
        mr: { xs: "auto", sm: 0.2 },
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

const NavBar: React.FC = () => {
  const auth = useContext(AuthContext);

  const userName = auth?.state.user ?? "";
  const userInfoQuery = useUserInfo(userName, {
    enabled: userName !== "",
  });

  const isDeveloper =
    userInfoQuery.isSuccess && userInfoQuery.data.user?.isDeveloper;
  const userMenu = UserMenu();
  const elements = (
    <>
      <Button color="inherit" href="/submissions">
        Submissions
      </Button>
      <Button color="inherit" href="/ranking">
        Ranking
      </Button>
      <Button color="inherit" href="/help">
        Help
      </Button>
      {isDeveloper && <ToolsMenu />}
      <LangSelect />
      {userMenu}
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

  const title = (
    <>
      <Button
        color="inherit"
        href="/"
        sx={{
          fontSize: "16px",
          fontWeight: "bolder",
        }}
      >
        Library Checker
      </Button>
    </>
  );

  const [open, setOpen] = React.useState(false);

  const toggleDrawerOpen = () => {
    setOpen(!open);
  };

  return (
    <>
      <AppBar position="static">
        <Toolbar sx={{ display: { xs: "none", sm: "flex" } }}>
          {title}
          {elements}
        </Toolbar>

        <Toolbar sx={{ display: { xs: "flex", sm: "none" } }}>
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
          display: { xs: "flex", sm: "none" },
        }}
        PaperProps={{
          sx: { width: "35%" },
        }}
        anchor="left"
        open={open}
        onClose={toggleDrawerOpen}
      >
        {elements}
      </Drawer>
    </>
  );
};

export default NavBar;
