import { Container, Typography } from "@material-ui/core";
import React from "react";
import AceEditor from "react-ace";
import { connect, PromiseState } from "react-refetch";
import { RouteComponentProps } from "react-router-dom";
import library_checker_client from "../api/library_checker_client";
import {
  SubmissionInfoRequest,
  SubmissionInfoResponse
} from "../api/library_checker_pb";
import SubmissionList from "../components/SubmissionList";

import "ace-builds/src-min-noconflict/mode-text";
import "ace-builds/src-min-noconflict/mode-c_cpp";

interface Props {
  submissionInfoFetch: PromiseState<SubmissionInfoResponse>;
}

const SubmissionInfo: React.FC<Props> = props => {
  const { submissionInfoFetch } = props;

  if (submissionInfoFetch.pending) {
    return <h1>Loading</h1>;
  }
  if (submissionInfoFetch.rejected) {
    return <h1>Error</h1>;
  }
  const info = submissionInfoFetch.value;
  const overview = info.getOverview()!;
  const lang = overview.getLang()!;

  const aceMode = (() => {
    if (lang.startsWith("cpp")) {
      return "c_cpp";
    }
    return "text";
  })();

  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Submission Info #{overview?.getId()}
      </Typography>
      {overview && <SubmissionList submissionOverviews={[overview]} />}
      <AceEditor
        mode={aceMode}
        value={info.getSource()}
        maxLines={10000}
        readOnly={true}
        width="100%"
        showPrintMargin={false}
      />
    </Container>
  );
};

export default connect<RouteComponentProps<{ submissionId: string }>, Props>(
  props => ({
    submissionInfoFetch: {
      comparison: null,
      value: library_checker_client.submissionInfo(
        new SubmissionInfoRequest().setId(
          parseInt(props.match.params.submissionId)
        )
      )
    }
  })
)(SubmissionInfo);
