## SWIM 

많은 분산 P2P(peer-to-peer) 어플리케이션은 모든 참여하는 Process에 대해 weakly-consistent한 Process 그룹 구성원 정보가 필요하다. SWIM은 대규모 프로세스 그룹에서 그룹 구성원 정보 서비스를 제공하는 범용 소프트웨어 모듈이다. SWIM은 전통적인 heart-beating 프로토콜의 unscalability를 극복하는 것을 목표로 한다. 전통적인 heart-beating protocol과는 다르게,  SWIM은 membership protocol에서 failure detection과 membership 업데이트 기능을 분리하였다.



### Basic SWIM Approach

SWIM은 크게 2개의 컴포넌트로 구성된다.

- A Failure Detector Component: 멤버들의 `failure`를 감지
- Dissemination Component: 최근에 `fail`, ` join`, or `left`한 멤버들의 정보를 전파



### SWIM Failure Detector

SWIM failure Detector 알고리즘은 2개의 파라미터를 사용한다: `T(protocol period)` 와 `k(the size of failure detection subgroups)` 

<p align="center">
<img src="../images/[swim]fig1.png" width="450px" height="400px">
</p>

`T` 마다 Failure Detection을 **Figure 1** 같이 수행한다. **Figure 1**은 임의의 노드 `M_i`를 대상으로 Failure Detection 알고리즘을 보여준다. 

1. `M_i` 는 `T`시간 마다 membership List에서 랜덤으로 member(`M_j`)를 하나 고르고, `M_j` 에게 `ping`을 보낸다. 

2. `M_i`가 ack timeout 시간(`T` 보다 작은 시간) 동안 `M_j`의 `ack`를 기다린다.
   1. `ack` 메세지가 올 경우 종료
   2. `ack` 메세지가 오지 않을 경우 3 수행 
3. `M_i`는 mebership List에서 `k`개의 member를 고르고, `ping-req(M_j)`를 보낸다.
4. `ping-req(M_j)`메세지를 받은 노드들은 `M_j`에게 `ping`을 보내고, `ack`를 받으면 `M_i`에게 전달한다.
5. `T` 가 끝날때 쯤, `M_i`는 `M_j`로 부터 혹은 `k`개의 member로 부터 `ack`메세지를 받았는지 확인하고, 메세지가 없을 경우 `M_j`는 fail되었다 판단하여 Dissemination component에게 memberlist update를 요청한다.



ack timeout은 네트워크 안에서의 round-trip을 기반으로 결정한다(평균 혹은 99%를 커버 하도록), `T`는 최소 round-trip보다 3배 이상 이어야 한다(논문에서는 평균 round-trip시간을 timeout으로 설정하였고, `T`는 평균 round-trip시간보다 훨씬 높은 값으로 설정하였다)

각 Message의 data는 unique sequence number를 가지고 있다. `ping`, `ping-req`, `ack` 메세지의 사이즈는 constant며 group의 size와는 관련이 없다. 

`M_i`가 `M_j`에게 `k`번의 `ping`을 날리지 않고 indirect하게 `M_i`의 다른 member에게 요청하는 이유는, `M_i`와 `M_j`사이의 네트워크의 혼잡성을 피하기 위함이다.  



### Dissemination Component and Dynamic Membership

Group member(`M_j`)의 fail을 감지하면, 노드(`M_i1`)는 Group의 다른 member에게 `failed(M_j)` 를 multicast한다. `failed(M_j)`를 받은 member들은 local membership List에서 `M_j`를 제거한다. 

새로운 노드가 들어오거나 자발적으로 나가게 되면, 해당 정보를 비슷한 방식으로 multicast한다. 그러나 어떤 노드가 Group에 들어오기 위해서는 그룹의 최소한 하나의 member를 알아야 한다. 그 방법은 다음과 같다.

- Group이 well known server와 관련이 있으면, 모든 `join`은 해당 address를 통해 이루어 질 수 있다.
- `Join` 메세지를 broadcast하고 각 group의 member들은 확률적으로 `reply` 한다.
- Group별로`Join`메세지를 전용으로 처리하는 Static coordinator를 둔다.