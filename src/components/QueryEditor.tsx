import React, { useState } from 'react';
import { CodeEditor, Button } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

import prettier from 'prettier';
import sqlPlugin from 'prettier-plugin-sql';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const [isExecuting, setIsExecuting] = useState(false);

  const onQueryTextChange = (newText: string) => {
    onChange({ ...query, queryText: newText });
  };

  const executeQuery = async () => {
    if (!query.queryText) {
      return;
    }
    setIsExecuting(true);
    try {
      await onRunQuery();
    } finally {
      setIsExecuting(false);
    }
  };

  const onFormatQuery = async () => {
    if (!queryText?.trim()) {
      return;
    }

    try {
      const formattedQuery = await prettier.format(queryText, {
        parser: 'sql',
        plugins: [sqlPlugin],
        printWidth: 120,
        tabWidth: 2,
        useTabs: false,
        singleQuote: true,
      });
      onQueryTextChange(formattedQuery);
    } catch (error) {
      alert(
        'SQL formatting failed. Please check if the syntax is correct.\nError message: ' + (error as Error).message
      );
    }
  };

  const { queryText } = query;

  return (
    <div data-testid="editor-container">
      <CodeEditor
        language="sql"
        height="200px"
        data-testid="editor-input"
        value={queryText || ''}
        showLineNumbers
        onChange={(newText) => onQueryTextChange(newText)}
      />
      <div style={{ marginTop: 8 }}>
        <Button 
          data-testid="run-query-button"
          onClick={executeQuery} disabled={isExecuting || !queryText}>
          {isExecuting ? 'Executing...' : 'Run Query'}
        </Button>
        <Button
          variant="secondary"
          data-testid="format-sql-button"
          onClick={onFormatQuery}
          disabled={isExecuting || !queryText}
          style={{ marginLeft: 8 }}
        >
          Format SQL
        </Button>
      </div>
    </div>
  );
}
