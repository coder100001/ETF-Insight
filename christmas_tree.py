import matplotlib.pyplot as plt
import numpy as np
from matplotlib.colors import LinearSegmentedColormap

# 常量定义
TREE_HEIGHT = 25
TREE_LAYERS = 126
ANGLE_DIVISIONS = 21
TREE_CROWN_START = 15
ANIMATION_FRAMES = 100

def create_tree_surface():
    """生成圣诞树主体的3D表面数据"""
    # 生成半径数据
    t = np.arange(start=0, stop=TREE_HEIGHT + 0.2, step=0.2)
    t[(t > 0) & (t <= 3)] = 1.5
    t[t > 3] = 8 - (t[t > 3] - 3) * 0.3636
    
    # 生成锥形曲面
    theta = np.linspace(0, 2 * np.pi, ANGLE_DIVISIONS)
    X = np.cos(theta).reshape((1, ANGLE_DIVISIONS))
    Y = np.sin(theta).reshape((1, ANGLE_DIVISIONS))
    Z = np.linspace(0, 1, TREE_LAYERS).reshape((TREE_LAYERS, 1))
    t = np.array(t).reshape((TREE_LAYERS, 1))
    
    X = X * t
    Y = Y * t
    Z = Z * np.ones_like(X) * TREE_HEIGHT
    
    # 随机移动树冠上点的位置，增加自然感
    crown_layers = TREE_LAYERS - TREE_CROWN_START
    angle = np.arctan(Y[TREE_CROWN_START:TREE_LAYERS, 0:ANGLE_DIVISIONS] / 
                      X[TREE_CROWN_START:TREE_LAYERS, 0:ANGLE_DIVISIONS])
    tree_diffusion = np.random.rand(crown_layers, ANGLE_DIVISIONS) - 0.5
    
    X[TREE_CROWN_START:TREE_LAYERS, 0:ANGLE_DIVISIONS] += np.cos(angle) * tree_diffusion
    Y[TREE_CROWN_START:TREE_LAYERS, 0:ANGLE_DIVISIONS] += np.sin(angle) * tree_diffusion
    Z[TREE_CROWN_START:TREE_LAYERS, 0:ANGLE_DIVISIONS] += (np.random.rand(crown_layers, ANGLE_DIVISIONS) - 0.5) * 0.5
    
    # 闭合曲面
    X[:, -1] = X[:, 0]
    Y[:, -1] = Y[:, 0]
    Z[:, -1] = Z[:, 0]
    
    return X, Y, Z

def create_tree_colormap():
    """创建圣诞树专用的绿色渐变色谱"""
    r = np.arange(start=0.0430, stop=0.2492, step=0.2061 / 50)
    g = np.arange(start=0.2969, stop=0.6982, step=0.4012 / 50)
    b = np.arange(start=0.0625, stop=0.3322, step=0.2696 / 50)
    
    rgb = np.column_stack([r, g, b])
    # 树干颜色（棕色）
    rgb[0:6, 0] = 77 / 265
    rgb[0:6, 1] = 63 / 265
    rgb[0:6, 2] = 5 / 265
    
    return LinearSegmentedColormap.from_list('tree_green', rgb)

def draw_tree_star(ax):
    """绘制圣诞树顶部的星星"""
    star_color = "#FFDF99"
    ax.scatter(0, 0, TREE_HEIGHT + 0.6, marker="*", c=star_color, s=500)
    ax.scatter(0, 0, TREE_HEIGHT, marker="o", c=star_color, s=7000, alpha=0.1)
class ChristmasLight:
    """圣诞彩灯类，用于管理彩灯的位置和绘制"""
    def __init__(self, z_range, num_points, radius_scale, angle_coef, color, marker, size):
        self.h = TREE_HEIGHT
        self.r = 8
        self.z = np.linspace(z_range[0], z_range[1], num_points)
        self.angle = angle_coef * np.pi
        self.radius_scale = radius_scale
        self.color = color
        self.marker = marker
        self.size = size
        
        # 预计算坐标
        self.x = self._calc_x(self.z)
        self.y = self._calc_y(self.z)
        self.handles = []
    
    def _calc_x(self, z):
        return (self.h - z) / self.h * self.r * self.radius_scale * np.cos(self.angle * z)
    
    def _calc_y(self, z):
        return (self.h - z) / self.h * self.r * self.radius_scale * np.sin(self.angle * z)
    
    def update_visibility(self, azim):
        """根据视角更新可见性"""
        visible_x = self.x.copy()
        visible_x[self.x * np.cos(azim) + self.y * np.sin(azim) < -2.5] = np.nan
        return visible_x
    
    def draw(self, ax, azim):
        """绘制彩灯"""
        visible_x = self.update_visibility(azim)
        if isinstance(self.size, list):
            # 多层光晕效果
            for i, s in enumerate(self.size):
                alpha = 0.8 if i == 0 else 0.05
                handle = ax.scatter(visible_x, self.y, self.z + 0.1, 
                                   marker=self.marker, c=self.color, 
                                   s=s, alpha=alpha, edgecolors='none')
                self.handles.append(handle)
        else:
            handle = ax.scatter(visible_x, self.y, self.z + 0.1, 
                               marker=self.marker, c=self.color, 
                               s=self.size, alpha=0.8, edgecolors='none')
            self.handles.append(handle)
    
    def remove(self):
        """移除当前绘制的彩灯"""
        for handle in self.handles:
            handle.remove()
        self.handles = []

def create_lights():
    """创建所有彩灯对象"""
    lights = []
    # 黄色彩灯组1 - 小点
    lights.append(ChristmasLight((4, TREE_HEIGHT - 4), 300, 1.5, 0.3, "#FDF9DC", ".", 5))
    # 黄色彩灯组2 - 大灯泡
    lights.append(ChristmasLight((4, TREE_HEIGHT - 4), 45, 1.5, 0.3, "#FDF9DC", "o", [60, 400]))
    # 白色彩灯组1 - 小点
    lights.append(ChristmasLight((4, TREE_HEIGHT - 6), 200, 1.45, -0.35, "white", ".", 5))
    # 白色彩灯组2 - 大灯泡
    lights.append(ChristmasLight((4, TREE_HEIGHT - 6), 17, 1.45, -0.35, "white", "o", [60, 400]))
    return lights

def draw_present(ax, dx, dy, dz, scale_x, scale_y, scale_z):
    """绘制礼物盒子"""
    present_x = np.array([[0.5, 0.5, 0.5, 0.5, 0.5],
                          [0, 1, 1, 0, 0],
                          [0, 1, 1, 0, 0],
                          [0, 1, 1, 0, 0],
                          [0.5, 0.5, 0.5, 0.5, 0.5]])
    present_y = np.array([[0.5, 0.5, 0.5, 0.5, 0.5],
                          [0, 0, 1, 1, 0],
                          [0, 0, 1, 1, 0],
                          [0, 0, 1, 1, 0],
                          [0.5, 0.5, 0.5, 0.5, 0.5]])
    present_z = np.array([[0, 0, 0, 0, 0],
                          [0, 0, 0, 0, 0],
                          [0.5, 0.5, 0.5, 0.5, 0.5],
                          [1, 1, 1, 1, 1],
                          [1, 1, 1, 1, 1]])
    ax.plot_surface(present_x * scale_x + dx, 
                    present_y * scale_y + dy, 
                    present_z * scale_z + dz)

def draw_presents(ax):
    """绘制所有礼物盒"""
    presents = [
        (-4, 4, 0, 2, 3, 1.5),
        (5, 3, 0, 4, 3, 3),
        (-7, -5, 0, 5, 3, 1),
        (-9, -6, 0, 2, 2, 2),
        (0, 7, 0, 4, 3, 3)
    ]
    for dx, dy, dz, sx, sy, sz in presents:
        draw_present(ax, dx, dy, dz, sx, sy, sz)


class Snow:
    """雪花类，管理雪花的位置和动画"""
    def __init__(self, num_particles, marker, size, alpha, fall_speed):
        self.positions = np.random.rand(num_particles, 3)
        self.positions[:, 0] = self.positions[:, 0] * 26 - 13
        self.positions[:, 1] = self.positions[:, 1] * 26 - 13
        self.positions[:, 2] = self.positions[:, 2] * 30
        self.marker = marker
        self.size = size
        self.alpha = alpha
        self.fall_speed = fall_speed
        self.handle = None
    
    def update(self):
        """更新雪花位置"""
        self.positions[:, 2] -= self.fall_speed
        # 重置落到底部的雪花
        self.positions[self.positions[:, 2] < 0, 2] = 30
    
    def draw(self, ax):
        """绘制雪花"""
        if self.handle:
            self.handle.remove()
        self.handle = ax.scatter(self.positions[:, 0], 
                                self.positions[:, 1], 
                                self.positions[:, 2], 
                                marker=self.marker, c="white", 
                                s=self.size, alpha=self.alpha, 
                                edgecolors='none')

def create_snow():
    """创建雪花对象"""
    return [
        Snow(40, ".", 20, 0.3, 0.25),
        Snow(20, "$*$", 80, 0.9, 0.5)
    ]

def setup_scene(ax):
    """设置场景样式和参数"""
    plt.rcParams['font.sans-serif'] = ['Cambria']
    plt.rcParams['axes.unicode_minus'] = False
    plt.rcParams['font.size'] = 20
    plt.rcParams['text.color'] = 'white'
    
    ax.text(0, 0, 30, 'Christmas Tree By slandarer', ha='center')
    ax.set_box_aspect((1, 1, 1))
    ax.set_position((-0.15, -0.2, 1.3, 1.3))
    ax.set_xlim(-10, 10)
    ax.set_ylim(-10, 10)
    ax.set_zlim(0, 30)
    ax.set_facecolor("#162033")
    ax.axis('off')
    ax.view_init(elev=10, azim=0)

def animate(ax, lights, snow_list):
    """运行动画循环"""
    for i in range(ANIMATION_FRAMES):
        # 更新雪花
        for snow in snow_list:
            snow.update()
            snow.draw(ax)
        
        # 更新彩灯
        azim = i / 180 * np.pi
        for light in lights:
            light.remove()
            light.draw(ax, azim)
        
        # 更新视角
        ax.view_init(elev=10, azim=i)
        plt.draw()
        plt.pause(0.1)

def main():
    """主函数"""
    # 创建3D图形
    fig, ax = plt.subplots(subplot_kw={"projection": "3d"})
    
    # 绘制圣诞树主体
    X, Y, Z = create_tree_surface()
    tree_colormap = create_tree_colormap()
    ax.plot_surface(X, Y, Z, cmap=tree_colormap)
    
    # 绘制星星
    draw_tree_star(ax)
    
    # 创建并绘制彩灯
    lights = create_lights()
    for light in lights:
        light.draw(ax, 0)
    
    # 绘制礼物盒
    draw_presents(ax)
    
    # 创建并绘制雪花
    snow_list = create_snow()
    for snow in snow_list:
        snow.draw(ax)
    
    # 设置场景
    setup_scene(ax)
    
    # 运行动画
    animate(ax, lights, snow_list)

if __name__ == "__main__":
    main()