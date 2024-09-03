import Typography from "@mui/material/Typography";
import React from "react";
import JudgeStatusList from "../components/JudgeStatusList";
import LangList from "../components/LangList";
import MainContainer from "../components/MainContainer";
import {
  Chip,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
} from "@mui/material";

const Help: React.FC = () => {
  return (
    <MainContainer title="Help">
      <>
        <Typography variant="h3" paragraph={true}>
          Terms and conditions
        </Typography>
        <Typography variant="body1" paragraph={true}>
          Users retain ownership of all intellectual and industrial property
          rights (including moral rights) in and to Submissions.
        </Typography>
        <Typography variant="body1" paragraph={true}>
          As a condition of submission, User grants the organizer a license to
          use, reproduce, adapt, modify, publish, distribute, publicly perform,
          create a derivative work from, and publicly display the Submission.
        </Typography>
        <Typography variant="body1" paragraph={true}>
          This license is based on{" "}
          <a href="https://opensource.google/docs/hackathons/#judge">
            Google Open Source Docs
          </a>
          , with some modifications.
        </Typography>
      </>

      <>
        <Typography variant="h3" paragraph={true}>
          Resource information
        </Typography>
        <List>
          <ListItem>
            <ListItemIcon sx={{ paddingRight: 1 }}>
              <Chip sx={{ width: 100 }} label="Server" />
            </ListItemIcon>
            <ListItemText primary="c2d-highcpu-8 (GCP)" />
          </ListItem>
          <ListItem>
            <ListItemIcon sx={{ paddingRight: 1 }}>
              <Chip sx={{ width: 100 }} label="CPU" />
            </ListItemIcon>
            <ListItemText primary="AMD EPYC™ 7B13 (limit 1 core)" />
          </ListItem>
          <ListItem>
            <ListItemIcon sx={{ paddingRight: 1 }}>
              <Chip sx={{ width: 100 }} label="Memory" />
            </ListItemIcon>
            <ListItemText primary="Limit 1 GiB" />
          </ListItem>
          <ListItem>
            <ListItemIcon sx={{ paddingRight: 1 }}>
              <Chip sx={{ width: 100 }} label="Stack Size" />
            </ListItemIcon>
            <ListItemText primary="Unlimited" />
          </ListItem>
        </List>
      </>

      <>
        <Typography variant="h3">Compiler list</Typography>
        <LangList />
        <Typography variant="body1" paragraph={true}>
          より詳しくは
          <a href="https://github.com/yosupo06/library-checker-judge/blob/master/langs/langs.toml">
            langs.toml
          </a>
          ,
          <a href="https://github.com/yosupo06/library-checker-judge/blob/master/langs/">
            Dockerfile
          </a>
          を参照してください
        </Typography>
      </>

      <>
        <Typography variant="h3">Judge status</Typography>
        <JudgeStatusList />
      </>
    </MainContainer>
  );
};

export default Help;
