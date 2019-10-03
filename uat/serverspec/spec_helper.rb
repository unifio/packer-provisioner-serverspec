require 'serverspec'
require 'net/ssh'

options = Net::SSH::Config.for(host, [])
options[:user] = ENV['TARGET_USER']
options[:keys] = ENV['TARGET_KEY']
options[:host_name] = ENV['TARGET_HOST']
options[:port] = ENV['TARGET_PORT']
options[:verify_host_key] = :never unless ENV['SERVERSPEC_HOST_KEY_CHECKING'] =~ (/^(true|t|yes|y|1)$/i)
backend = ENV.fetch('SERVERSPEC_BACKEND', 'ssh').to_sym

set :backend,      backend
set :disable_sudo, true
set :host,         options[:host_name]
set :path,         '/usr/local/bin:$PATH'
set :request_pty,  true
set :ssh_options,  options
