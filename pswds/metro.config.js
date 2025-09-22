const {getDefaultConfig} = require('expo/metro-config');
const {mergeConfig} = require('@react-native/metro-config');
const {
  wrapWithReanimatedMetroConfig,
} = require('react-native-reanimated/metro-config');

/**
 * Metro configuration
 * https://reactnative.dev/docs/metro
 *
 * @type {import('metro-config').MetroConfig}
 */
const defaultConfig = getDefaultConfig(__dirname);
defaultConfig.resolver.sourceExts.push('sql');

module.exports = mergeConfig(
  defaultConfig,
  wrapWithReanimatedMetroConfig({
    resolver: {
      unstable_enablePackageExports: true,
    },
  }),
);
