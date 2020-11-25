import { AppBar, Button, List, ListItem, ListItemText, MenuItem, Select, Toolbar, Typography } from '@material-ui/core';
import React, { useContext } from 'react';
import { RouteComponentProps, withRouter } from 'react-router-dom';
import { LangContext } from '../contexts/LangContext';
import { AuthContext } from '../contexts/AuthContext';

const NavBar = (props: RouteComponentProps) => {
  const { history } = props
  const lang = useContext(LangContext)
  const auth = useContext(AuthContext)
  return (
    <AppBar position="static">
      <Toolbar>
        <List>
          <ListItem>
            <ListItemText>
              <Typography color="inherit" variant="h6">
                <Button color="inherit" onClick={() => history.push('/')}>
                  LIBRARY-CHECKER
                </Button>
              </Typography>
            </ListItemText>
            <ListItemText inset>
              <Typography color="inherit" variant="h6">
                <Button color="inherit" onClick={() => history.push('/submissions')}>
                  SUBMISSIONS
                </Button>
              </Typography>
            </ListItemText>
            <ListItemText>
              <Typography color="inherit" variant="h6">
                <Button color="inherit" onClick={() => history.push('/ranking')}>
                  RANKING
                </Button>
              </Typography>
            </ListItemText>
            <ListItemText>
              <Typography color="inherit" variant="h6">
                <Button color="inherit" onClick={() => history.push('/help')}>
                  HELP
                </Button>
              </Typography>
            </ListItemText>
          </ListItem>
        </List> 
        <Select          
          value={lang?.state.lang}
          onChange={(e) => lang?.dispatch({ type: 'change', payload: e.target.value as string === 'ja' ? 'ja' : 'en'})}
          style={{
            marginLeft: 'auto',
            minWidth: 120,
          }}
        >
          <MenuItem value='ja'>Japanese</MenuItem>
          <MenuItem value='en'>English</MenuItem>
        </Select>
        {
          !auth?.state.user && 
          <Button color="inherit" onClick={() => history.push('/login')}>
            Login
          </Button>
        }
        {
          auth?.state.user &&
          <Typography color="inherit" variant="h6">
            {auth?.state.user}
          </Typography>
        }

      </Toolbar>
    </AppBar>
  );
};

export default withRouter(NavBar);
