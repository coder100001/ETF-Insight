"""
清理旧的无步骤工作流实例
由于之前的工作流执行没有正确记录步骤，这些实例对新的工作流系统没有用处
"""

from workflow.models import WorkflowInstance, WorkflowInstanceStep

def cleanup_old_instances():
    """清理没有步骤实例记录的旧实例"""
    print("开始清理旧的无步骤实例...")

    # 找出所有工作流实例
    all_instances = WorkflowInstance.objects.all()
    print(f"当前共有 {all_instances.count()} 个工作流实例")

    # 统计每个实例的步骤数
    no_step_instances = []
    for inst in all_instances:
        step_count = inst.step_instances.count()
        if step_count == 0:
            no_step_instances.append(inst)

    print(f"其中 {len(no_step_instances)} 个实例没有步骤记录")

    if len(no_step_instances) > 0:
        # 删除这些实例
        confirm = input(f"\n是否删除这 {len(no_step_instances)} 个无步骤实例？(yes/no): ")
        if confirm.lower() == 'yes':
            count = 0
            for inst in no_step_instances:
                print(f"  删除实例 #{inst.id}: {inst.workflow.name}")
                inst.delete()
                count += 1
            print(f"\n已删除 {count} 个无步骤实例")
        else:
            print("跳过删除操作")
    else:
        print("没有需要清理的实例")

    # 显示清理后的统计
    print(f"\n清理后统计:")
    print(f"  工作流实例总数: {WorkflowInstance.objects.count()}")
    print(f"  实例步骤记录总数: {WorkflowInstanceStep.objects.count()}")

    # 按工作流统计
    from workflow.models import Workflow
    for wf in Workflow.objects.all().order_by('id'):
        total_inst = wf.instances.count()
        total_steps = WorkflowInstanceStep.objects.filter(
            workflow_instance__workflow=wf
        ).count()
        print(f"  {wf.name}: {total_inst}个实例, {total_steps}个步骤记录")


if __name__ == '__main__':
    cleanup_old_instances()
