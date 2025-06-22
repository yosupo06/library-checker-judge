import Button from "@mui/material/Button";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import Menu from "@mui/material/Menu";
import { GitHub } from "@mui/icons-material";
import MenuIcon from "@mui/icons-material/Menu";
import { SvgIcon } from "@mui/material";
import React from "react";
import { Link, LinkProps, useNavigate } from "react-router-dom";
import { useLangContext } from "../../contexts/LangContext";
import flagJA from "./flag_ja.svg";
import flagUS from "./flag_us.svg";
import { styled } from "@mui/material/styles";
import { useCurrentUser } from "../../api/client_wrapper";
import { Drawer, IconButton } from "@mui/material";
import { useSignOutMutation } from "../../auth/auth";

const ButtonLink = styled(Button)<LinkProps>();

// https://github.com/mui/material-ui/issues/35218#issuecomment-1977984142
const DiscordIcon: React.FC = () => (
  <SvgIcon>
    <path d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.9555 2.4189-2.1569 2.4189zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.4189-2.1568 2.4189Z" />
  </SvgIcon>
);

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
      <ButtonLink LinkComponent={Link} color="inherit" to="/hacks">
        Hacks
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
        href="https://discord.gg/DfJvTe5tzD"
        target="_blank"
        rel="noopener"
      >
        <DiscordIcon />
      </Button>
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
        <MenuItem onClick={handleClose}>
          <NavbarLink to={`/tool/statementviewer`}>Statement Viewer</NavbarLink>
        </MenuItem>
        <MenuItem onClick={handleClose}>
          <NavbarLink to={`/tool/monitoring`}>Monitoring</NavbarLink>
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
