import { css } from '@emotion/css';

import { GrafanaTheme2 } from '@grafana/data';
import { useStyles2 } from '@grafana/ui';
import React from 'react';

export function CheatSheet() {
  const styles = useStyles2(getStyles);

  return (
    <div>
      <h2>KWDB cheat sheet</h2>
       Format SQL:
      <ul className={styles.ulPadding}>
        <li>
          You can use the <code>Format SQL</code> button to format your SQL queries.
        </li>
      </ul>
       Macros:
      <ul className={styles.ulPadding}>
        <li>$from -&gt; From in grafana time range</li>
        <li>$to -&gt; To in grafana time range</li>
        <li>$interval -&gt; Time interval in grafana time range</li>
      </ul>
    </div>
  );
}

function getStyles(theme: GrafanaTheme2) {
  return {
    ulPadding: css({
      margin: theme.spacing(1, 0),
      paddingLeft: theme.spacing(5),
    }),
  };
}
