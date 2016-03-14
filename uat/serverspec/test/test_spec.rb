require_relative '../spec_helper'

describe package('ntp') do
  it { should be_installed }
end
