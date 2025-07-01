import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput, FieldSet, Alert, Icon } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        host: event.target.value,
      },
    });
  };

  const onPortChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        port: parseInt(event.target.value, 10),
      },
    });
  };

  const onDatabaseChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        database: event.target.value,
      },
    });
  };

  const onUsernameChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        username: event.target.value,
      },
    });
  };

  // Secure fields (only sent to the backend)
  const onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...secureJsonData,
        password: event.target.value,
      },
    });
  };

  const onResetPassword = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  return (
    <div className="gf-form-group">
      <Alert title="User Permission" severity="warning">
        <div style={{ lineHeight: '1.8' }}>
          <p>
            For security, grant the database user <b>ONLY SELECT</b> permissions on the specific databases and tables
            needed for queries, Grafana does not validate query safety, meaning any SQL command (e.g.,{' '}
            <code>DELETE FROM user;</code>, <code>DROP TABLE user;</code>) will be executed. To mitigate this risk,
            strongly recommend creating a dedicated database user with strictly <code>SELECT</code> privileges.
            <br />
            Refer to the{' '}
            <a
              href="https://www.kaiwudb.com/kaiwudb_docs/#/oss_dev/db-security/privilege-mgmt.html"
              target="_blank"
              rel="noreferrer"
            >
              documentation
            </a>{' '}
            for details.
          </p>
        </div>
      </Alert>
      <div className="gf-form--v-stack">
        <FieldSet label="Connection">
          <InlineField
            label="Host URL"
            labelWidth={20}
            required
            tooltip="KWDB server address"
            interactive
            error="The host address cannot be empty"
            invalid={!jsonData.host}
          >
            <Input
              id="config-editor-host"
              name="host"
              data-testid="host-input"
              onChange={onHostChange}
              value={jsonData.host}
              placeholder="localhost"
              width={40}
              prefix={<Icon name="link" />}
            />
          </InlineField>

          <InlineField
            label="Port"
            labelWidth={20}
            required
            tooltip={
              <div>
                Default port 26257
                <br />
                Range 1-65535
              </div>
            }
            invalid={typeof jsonData.port === 'undefined' || jsonData.port < 1 || jsonData.port > 65535}
            error="Invalid port number"
          >
            <Input
              type="number"
              name="port"
              data-testid="port-input"
              min={1}
              max={65535}
              step={1}
              value={jsonData.port}
              onChange={onPortChange}
              width={40}
              placeholder="26257"
              prefix={<Icon name="compass" />}
            />
          </InlineField>

          <InlineField label="Database name" labelWidth={20} required tooltip="The default DB is defaultdb">
            <Input
              value={jsonData.database}
              data-testid="database-input"
              onChange={onDatabaseChange}
              placeholder="defaultdb"
              width={40}
              prefix={<Icon name="database" />}
            />
          </InlineField>
        </FieldSet>

        <FieldSet label="Authentication">
          <InlineField label="Username" labelWidth={20} required interactive tooltip="Database login account">
            <Input
              value={jsonData.username}
              data-testid="username-input"
              onChange={onUsernameChange}
              placeholder="root"
              width={40}
              prefix={<Icon name="user" />}
            />
          </InlineField>

          <InlineField label="Password" labelWidth={20} interactive>
            <SecretInput
              isConfigured={secureJsonFields.password}
              value={secureJsonData?.password}
              data-testid="password-input"
              onReset={onResetPassword}
              onChange={onPasswordChange}
              width={40}
              prefix={<Icon name="shield-exclamation" />}
            />
          </InlineField>
        </FieldSet>
      </div>
    </div>
  );
}
