export const TaskOrderStatus = {
  DRAFT: 0,
  SUBMITTED: 1,
  IN_PROGRESS: 2,
  COMPLETED: 3,
} as const

export type TaskOrderStatusValue = (typeof TaskOrderStatus)[keyof typeof TaskOrderStatus]

export const TaskOrderStatusLabel: Record<TaskOrderStatusValue, string> = {
  [TaskOrderStatus.DRAFT]: '草稿',
  [TaskOrderStatus.SUBMITTED]: '已提交',
  [TaskOrderStatus.IN_PROGRESS]: '流程中',
  [TaskOrderStatus.COMPLETED]: '已完成',
}

export const TaskOrderStatusTagType: Record<TaskOrderStatusValue, string> = {
  [TaskOrderStatus.DRAFT]: 'info',
  [TaskOrderStatus.SUBMITTED]: 'warning',
  [TaskOrderStatus.IN_PROGRESS]: 'primary',
  [TaskOrderStatus.COMPLETED]: 'success',
}

export const ReviewResult = {
  PASS: '通过',
  REJECT: '驳回',
  PENDING: '待评审',
} as const

export type ReviewResultValue = (typeof ReviewResult)[keyof typeof ReviewResult]

export const ReviewResultTagType: Record<string, string> = {
  [ReviewResult.PASS]: 'success',
  [ReviewResult.REJECT]: 'danger',
  [ReviewResult.PENDING]: 'info',
}

export const ApprovalAction = {
  APPROVE: 'approve',
  REJECT: 'reject',
} as const

export type ApprovalActionValue = (typeof ApprovalAction)[keyof typeof ApprovalAction]