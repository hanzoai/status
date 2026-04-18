import { getDefaultGuiConfig, createGui } from '@hanzo/gui'

const defaultConfig = getDefaultGuiConfig('web')

export default createGui({
  ...defaultConfig,
  settings: {
    ...defaultConfig.settings,
    shouldAddPrefersColorThemes: false,
  },
})
