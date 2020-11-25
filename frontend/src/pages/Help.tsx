import { Container, Typography } from '@material-ui/core'
import React from 'react'
import { connect } from 'react-refetch'
import JudgeStatusList from '../components/JudgeStatusList'
import LangList from '../components/LangList'

interface Props {
}

const Help: React.FC<Props> = (props) => {
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>Help</Typography>
      <Typography variant="h3" paragraph={true}>Terms and conditions</Typography>
      <Typography variant="body1" paragraph={true}>
        Users retain ownership of all intellectual and industrial property rights (including moral rights) in and to Submissions.
      </Typography>
      <Typography variant="body1" paragraph={true}>
        As a condition of submission, User grants the organizer a license to use, reproduce, adapt, modify, publish, distribute, publicly perform, create a derivative work from, and publicly
        display the Submission.
      </Typography>
      <Typography variant="body1" paragraph={true}>
        This license is based on <a href="https://opensource.google/docs/hackathons/#judge">Google Open Source Docs</a>, with some modifications
      </Typography>

      <Typography variant="h3">Lang List</Typography>
      <LangList />
      <Typography variant="body1" paragraph={true}>
        より詳しくは
        <a href="https://github.com/yosupo06/library-checker-judge/blob/master/api/langs.toml">langs.toml</a>,
        <a href="https://github.com/yosupo06/library-checker-judge/blob/master/deploy/install.sh">install.sh</a>を参照してください
      </Typography>

      <Typography variant="h3">Judge Status</Typography>
      <JudgeStatusList />

      <Typography variant="h3">Tips</Typography>
      <Typography variant="body1" paragraph={true}>
        Memory Limit is an 1G for all problems. Stack Size Limit is unlimited.
      </Typography>
      <Typography variant="body1" paragraph={true}>
        We will restart judge servers sometimes. If you submit your solution while restarting, it may take a longer time (~ 5min) rather than usual.
      </Typography>


    </Container>
  );
}

export default connect<{}, Props>(() => ({}))(Help);
