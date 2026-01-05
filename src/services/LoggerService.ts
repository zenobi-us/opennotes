import { pino } from 'pino';

export const Logger = pino({
  name: 'wiki',
  level: process.env.LOG_LEVEL || 'info',
  enabled: !!process.env.DEBUG,
  transport: {
    target: 'pino-pretty',
    options: {
      colorize: true,
    },
  },
});
