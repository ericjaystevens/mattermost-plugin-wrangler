package main

import (
	"io/ioutil"
	"testing"

	"github.com/ericjaystevens/slashparse"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	context := &plugin.Context{}

	commandUser := &model.User{
		Id:    model.NewId(),
		Email: "user@emaildomain.com",
	}

	api := &plugintest.API{}
	api.On("GetUser", commandUser.Id).Return(commandUser, nil)
	api.On("GetUser", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil, &model.AppError{DetailedError: "invalid user"})

	var plugin Plugin
	plugin.SetAPI(api)
	configPath := `/home/ec2-user/code/mattermost-plugin-wrangler/wrangler.yaml`
	slashDef, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Error()
	}
	plugin.slashCommand, _ = slashparse.NewSlashCommand(slashDef)

	t.Run("args", func(t *testing.T) {
		t.Run("no args", func(t *testing.T) {
			args := &model.CommandArgs{}
			resp, appErr := plugin.ExecuteCommand(context, args)
			require.Nil(t, appErr)
			require.Equal(t, resp.Text, getHelp())
		})

		t.Run("one arg", func(t *testing.T) {
			args := &model.CommandArgs{Command: "one"}
			resp, appErr := plugin.ExecuteCommand(context, args)
			require.Nil(t, appErr)
			require.Equal(t, "/one is not a valid command. Please see /wrangler help", resp.Text)
		})

		t.Run("two args, invalid command", func(t *testing.T) {
			args := &model.CommandArgs{Command: "one two"}
			resp, appErr := plugin.ExecuteCommand(context, args)
			require.Nil(t, appErr)
			require.Equal(t, "/one is not a valid command. Please see /wrangler help", resp.Text)
		})

		t.Run("move command", func(t *testing.T) {
			t.Run("missing extra args", func(t *testing.T) {
				args := &model.CommandArgs{Command: "wrangler move"}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				require.Equal(t, "/wrangler move requires an additional command. Try adding thread. Please see /wrangler help for more info", resp.Text)
			})

			t.Run("invalid extra args", func(t *testing.T) {
				args := &model.CommandArgs{Command: "wrangler move invalid"}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				require.Equal(t, "/wrangler move requires an additional command. Try adding thread. Please see /wrangler help for more info", resp.Text)
			})
		})

		t.Run("copy command", func(t *testing.T) {
			t.Run("missing extra args", func(t *testing.T) {
				args := &model.CommandArgs{Command: "wrangler copy"}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				require.Equal(t, "/wrangler copy requires an additional command. Try adding thread. Please see /wrangler help for more info", resp.Text)
			})

			t.Run("invalid extra args", func(t *testing.T) {
				args := &model.CommandArgs{Command: "wrangler copy invalid"}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				require.Equal(t, "/wrangler copy requires an additional command. Try adding thread. Please see /wrangler help for more info", resp.Text)
			})
		})

		t.Run("attach command", func(t *testing.T) {
			t.Run("missing extra args", func(t *testing.T) {
				args := &model.CommandArgs{Command: "wrangler attach"}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				require.Equal(t, "/wrangler attach requires an additional command. Try adding message. Please see /wrangler help for more info", resp.Text)
			})

			t.Run("invalid extra args", func(t *testing.T) {
				args := &model.CommandArgs{Command: "wrangler attach invalid"}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				require.Equal(t, "/wrangler attach requires an additional command. Try adding message. Please see /wrangler help for more info", resp.Text)
			})
		})

		t.Run("list command", func(t *testing.T) {
			t.Run("missing extra args", func(t *testing.T) {
				args := &model.CommandArgs{Command: "wrangler list"}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				require.Equal(t, "/wrangler list requires an additional command. Try adding channels or messages. Please see /wrangler help for more info", resp.Text)
			})

			t.Run("invalid extra args", func(t *testing.T) {
				args := &model.CommandArgs{Command: "wrangler list invalid"}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				require.Equal(t, "/wrangler list requires an additional command. Try adding channels or messages. Please see /wrangler help for more info", resp.Text)
			})
		})
	})

	t.Run("info command", func(t *testing.T) {
		args := &model.CommandArgs{Command: "wrangler info"}
		resp, appErr := plugin.ExecuteCommand(context, args)
		require.Nil(t, appErr)
		infoResp, userError, err := plugin.runInfoCommand([]string{}, nil)
		require.NoError(t, err)
		assert.False(t, userError)
		assert.Equal(t, infoResp, resp)
	})

	t.Run("allowed email domain", func(t *testing.T) {
		t.Run("enabled, user not in domain", func(t *testing.T) {
			plugin.setConfiguration(&configuration{
				AllowedEmailDomain: "baddomain.com",
			})
			args := &model.CommandArgs{
				UserId:  commandUser.Id,
				Command: "wrangler info",
			}
			resp, appErr := plugin.ExecuteCommand(context, args)
			require.Nil(t, appErr)
			assert.Equal(t, "Permission denied. Please talk to your system administrator to get access.", resp.Text)
		})

		t.Run("enabled, user in domain", func(t *testing.T) {
			plugin.setConfiguration(&configuration{
				AllowedEmailDomain: "emaildomain.com",
			})
			args := &model.CommandArgs{
				UserId:  commandUser.Id,
				Command: "wrangler info",
			}
			resp, appErr := plugin.ExecuteCommand(context, args)
			require.Nil(t, appErr)
			infoResp, userError, err := plugin.runInfoCommand([]string{}, nil)
			require.NoError(t, err)
			assert.False(t, userError)
			assert.Equal(t, resp, infoResp)
		})

		t.Run("enabled, invalid user", func(t *testing.T) {
			plugin.setConfiguration(&configuration{
				AllowedEmailDomain: "emaildomain.com",
			})
			args := &model.CommandArgs{
				UserId:  model.NewId(),
				Command: "wrangler info",
			}
			resp, appErr := plugin.ExecuteCommand(context, args)
			require.Nil(t, appErr)
			assert.Equal(t, "Permission denied. Please talk to your system administrator to get access.", resp.Text)
		})

		t.Run("multiple domains", func(t *testing.T) {
			t.Run("user in first domain", func(t *testing.T) {
				plugin.setConfiguration(&configuration{
					AllowedEmailDomain: "emaildomain.com,anotherdomain.com",
				})
				args := &model.CommandArgs{
					UserId:  commandUser.Id,
					Command: "wrangler info",
				}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				infoResp, userError, err := plugin.runInfoCommand([]string{}, nil)
				require.NoError(t, err)
				assert.False(t, userError)
				assert.Equal(t, infoResp, resp)
			})

			t.Run("user in second domain", func(t *testing.T) {
				commandUser.Email = "user@anotherdomain.com"
				plugin.setConfiguration(&configuration{
					AllowedEmailDomain: "emaildomain.com,anotherdomain.com",
				})
				args := &model.CommandArgs{
					UserId:  commandUser.Id,
					Command: "wrangler info",
				}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				infoResp, userError, err := plugin.runInfoCommand([]string{}, nil)
				require.NoError(t, err)
				assert.False(t, userError)
				assert.Equal(t, infoResp, resp)
			})

			t.Run("user in neither domain", func(t *testing.T) {
				commandUser.Email = "user@anotherbaddomain.com"
				plugin.setConfiguration(&configuration{
					AllowedEmailDomain: "emaildomain.com,anotherdomain.com",
				})
				args := &model.CommandArgs{
					UserId:  commandUser.Id,
					Command: "wrangler info",
				}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				assert.Equal(t, "Permission denied. Please talk to your system administrator to get access.", resp.Text)
			})

			t.Run("user is a direct email match", func(t *testing.T) {
				commandUser.Email = "user1@test.com"
				plugin.setConfiguration(&configuration{
					AllowedEmailDomain: "emaildomain.com,anotherdomain.com,user1@test.com",
				})
				args := &model.CommandArgs{
					UserId:  commandUser.Id,
					Command: "wrangler info",
				}
				resp, appErr := plugin.ExecuteCommand(context, args)
				require.Nil(t, appErr)
				infoResp, userError, err := plugin.runInfoCommand([]string{}, nil)
				require.NoError(t, err)
				assert.False(t, userError)
				assert.Equal(t, infoResp, resp)
			})
		})
	})
}
