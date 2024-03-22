import * as core from '@actions/core'
import { zip } from '@grafana/bundle-plugin'

/**
 * The main function for the action.
 * @returns {Promise<void>} Resolves when the action is complete.
 */
export async function run(): Promise<void> {
  try {
    const distDir: string = core.getInput('distDir')
    const outDir: string = core.getInput('outDir')
    zip({ distDir, outDir })
  } catch (error) {
    // Fail the workflow run if an error occurs
    if (error instanceof Error) core.setFailed(error.message)
  }
}
